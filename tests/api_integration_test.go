package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pleco-api/internal/appsetup"
	"pleco-api/internal/config"
	"pleco-api/internal/services"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestAPI_Login_Integration(t *testing.T) {
	// Setup mock DB
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	// Setup Test Config
	cfg := config.AppConfig{
		JWTSecret: []byte("test-secret-key-at-least-32-bytes-long-123456"),
	}
	jwtService := services.NewJWTService(cfg.JWTSecret)

	// Build Router
	router, err := appsetup.BuildRouter(db, cfg, jwtService, nil)
	require.NoError(t, err)

	// Define Test Data
	email := "test@example.com"
	password := "Password123!"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Mock DB Expectations for Login
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "role", "role_id", "is_verified", "password_updated_at", "access_token_version"}).
		AddRow(1, time.Now(), time.Now(), nil, "Test User", email, string(hashedPassword), "user", 1, true, time.Now(), 0)

	// Expect query to find user by email
	mock.ExpectQuery(`SELECT \* FROM "users".*WHERE email = \$1.*`).
		WithArgs(email, sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Expect save for refresh token
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "refresh_tokens" SET "deleted_at"=\$1 WHERE \(user_id = \$2 AND device_id = \$3\) AND "refresh_tokens"\."deleted_at" IS NULL`).
		WithArgs(sqlmock.AnyArg(), 1, "test-device").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`INSERT INTO "refresh_tokens".*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// Expect login activity timestamp update
	mock.ExpectExec(`UPDATE "users" SET "last_login_at"=\$1,"updated_at"=\$2 WHERE id = \$3 AND "users"\."deleted_at" IS NULL`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect save for audit log (SafeRecord uses a transaction)
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "audit_logs".*`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// Prepare Login Request
	loginData := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(loginData)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-ID", "test-device")

	// Perform Request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Login success", response["message"])

	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])
	assert.Nil(t, data["refresh_token"])
	assertRefreshCookie(t, w.Result().Cookies())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func assertRefreshCookie(t *testing.T, cookies []*http.Cookie) {
	t.Helper()

	for _, cookie := range cookies {
		if cookie.Name == "pleco_refresh_token" {
			assert.NotEmpty(t, cookie.Value)
			assert.True(t, cookie.HttpOnly)
			assert.True(t, cookie.Secure)
			assert.Equal(t, http.SameSiteNoneMode, cookie.SameSite)
			assert.Equal(t, "/", cookie.Path)
			return
		}
	}

	t.Fatal("expected pleco_refresh_token cookie")
}

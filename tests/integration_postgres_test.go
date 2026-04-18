package tests

import (
	"os"
	"testing"
	"time"

	"go-api-starterkit/modules/audit"
	"go-api-starterkit/modules/permission"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openPostgresIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	return db
}

func setupPermissionTempTables(t *testing.T, db *gorm.DB) *gorm.DB {
	t.Helper()

	tx := db.Begin()
	require.NoError(t, tx.Error)

	require.NoError(t, tx.Exec(`CREATE TEMP TABLE roles (id SERIAL PRIMARY KEY, name TEXT NOT NULL UNIQUE)`).Error)
	require.NoError(t, tx.Exec(`CREATE TEMP TABLE role_permissions (id SERIAL PRIMARY KEY, role_id INTEGER NOT NULL, permission VARCHAR(255) NOT NULL, created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP)`).Error)

	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})

	return tx
}

func setupAuditTempTable(t *testing.T, db *gorm.DB) *gorm.DB {
	t.Helper()

	tx := db.Begin()
	require.NoError(t, tx.Error)

	require.NoError(t, tx.Exec(`
		CREATE TEMP TABLE audit_logs (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE,
			actor_user_id INTEGER,
			action VARCHAR(100) NOT NULL,
			resource VARCHAR(100) NOT NULL,
			resource_id INTEGER,
			status VARCHAR(50) NOT NULL DEFAULT 'success',
			description TEXT,
			ip_address VARCHAR(255),
			user_agent TEXT
		)
	`).Error)

	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})

	return tx
}

func TestPermissionRepository_HasRolePermission_WithTempTables(t *testing.T) {
	db := openPostgresIntegrationDB(t)
	tx := setupPermissionTempTables(t, db)

	require.NoError(t, tx.Exec(`INSERT INTO roles (id, name) VALUES (1, 'admin')`).Error)
	require.NoError(t, tx.Exec(`INSERT INTO role_permissions (role_id, permission) VALUES (1, 'user.read_all')`).Error)

	repo := permission.NewRepository(tx)

	allowed, err := repo.HasRolePermission("admin", "user.read_all")
	require.NoError(t, err)
	assert.True(t, allowed)

	denied, err := repo.HasRolePermission("admin", "user.delete")
	require.NoError(t, err)
	assert.False(t, denied)
}

func TestAuditRepository_FindAllWithFilter_WithTempTable(t *testing.T) {
	db := openPostgresIntegrationDB(t)
	tx := setupAuditTempTable(t, db)

	now := time.Now()
	require.NoError(t, tx.Exec(`
		INSERT INTO audit_logs (created_at, updated_at, action, resource, status, description)
		VALUES
			(?, ?, 'login', 'auth', 'success', 'user logged in'),
			(?, ?, 'create_user', 'user', 'success', 'admin created user'),
			(?, ?, 'update_user', 'user', 'success', 'admin updated user')
	`, now.Add(-3*time.Minute), now.Add(-3*time.Minute), now.Add(-2*time.Minute), now.Add(-2*time.Minute), now.Add(-1*time.Minute), now.Add(-1*time.Minute)).Error)

	repo := audit.NewRepository(tx)

	logs, total, err := repo.FindAllWithFilter(1, 10, "", "user")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, logs, 2)
	assert.Equal(t, "update_user", logs[0].Action)
	assert.Equal(t, "create_user", logs[1].Action)
}

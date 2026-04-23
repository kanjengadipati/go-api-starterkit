package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	aiModule "go-api-starterkit/internal/ai"
	"go-api-starterkit/internal/config"
	"go-api-starterkit/internal/modules/audit"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type stubAuditRepo struct {
	findAllWithFilter     func(filter audit.Filter) ([]audit.AuditLog, int64, error)
	findForExport         func(filter audit.Filter) ([]audit.AuditLog, error)
	createLog             func(log *audit.AuditLog) error
	createInvestigation   func(record *audit.AuditInvestigation) error
	findBySnapshot        func(createdByUserID *uint, snapshotHash string) (*audit.AuditInvestigation, error)
	findInvestigations    func(filter audit.InvestigationFilter) ([]audit.AuditInvestigation, int64, error)
	findInvestigationByID func(id uint) (*audit.AuditInvestigation, error)
}

func (s *stubAuditRepo) Create(log *audit.AuditLog) error {
	if s.createLog != nil {
		return s.createLog(log)
	}
	return nil
}

func (s *stubAuditRepo) FindAllWithFilter(filter audit.Filter) ([]audit.AuditLog, int64, error) {
	if s.findAllWithFilter != nil {
		return s.findAllWithFilter(filter)
	}
	return nil, 0, nil
}

func (s *stubAuditRepo) FindForExport(filter audit.Filter) ([]audit.AuditLog, error) {
	if s.findForExport != nil {
		return s.findForExport(filter)
	}
	return nil, nil
}

func (s *stubAuditRepo) CreateInvestigation(record *audit.AuditInvestigation) error {
	if s.createInvestigation != nil {
		return s.createInvestigation(record)
	}
	record.ID = 1
	return nil
}

func (s *stubAuditRepo) FindLatestInvestigationBySnapshot(createdByUserID *uint, snapshotHash string) (*audit.AuditInvestigation, error) {
	if s.findBySnapshot != nil {
		return s.findBySnapshot(createdByUserID, snapshotHash)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubAuditRepo) FindInvestigations(filter audit.InvestigationFilter) ([]audit.AuditInvestigation, int64, error) {
	if s.findInvestigations != nil {
		return s.findInvestigations(filter)
	}
	return nil, 0, nil
}

func (s *stubAuditRepo) FindInvestigationByID(id uint) (*audit.AuditInvestigation, error) {
	if s.findInvestigationByID != nil {
		return s.findInvestigationByID(id)
	}
	return nil, gorm.ErrRecordNotFound
}

func TestAuditHandler_GetLogs_WithExtendedFilter(t *testing.T) {
	expectedActorID := uint(42)
	expectedFrom, _ := time.Parse(time.RFC3339, "2026-04-20T00:00:00Z")
	expectedTo, _ := time.Parse(time.RFC3339, "2026-04-21T00:00:00Z")

	service := audit.NewService(&stubAuditRepo{
		findAllWithFilter: func(filter audit.Filter) ([]audit.AuditLog, int64, error) {
			assert.Equal(t, 2, filter.Page)
			assert.Equal(t, 20, filter.Limit)
			assert.Equal(t, "login", filter.Action)
			assert.Equal(t, "auth", filter.Resource)
			assert.Equal(t, "failed", filter.Status)
			assert.Equal(t, "203.0.113.1", filter.Search)
			assert.NotNil(t, filter.ActorUserID)
			assert.Equal(t, expectedActorID, *filter.ActorUserID)
			assert.NotNil(t, filter.DateFrom)
			assert.NotNil(t, filter.DateTo)
			assert.True(t, filter.DateFrom.Equal(expectedFrom))
			assert.True(t, filter.DateTo.Equal(expectedTo))

			return []audit.AuditLog{{Action: "login", Resource: "auth", Status: "failed"}}, 1, nil
		},
	})

	handler := audit.NewHandler(service, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodGet,
		"/audit-logs?page=2&limit=20&action=login&resource=auth&status=failed&actor_user_id=42&search=203.0.113.1&date_from=2026-04-20T00:00:00Z&date_to=2026-04-21T00:00:00Z",
		nil,
	)

	handler.GetLogs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "success", bodyMap["status"])
	assert.Equal(t, "Audit logs fetched", bodyMap["message"])
	meta := bodyMap["meta"].(map[string]interface{})
	assert.Equal(t, float64(2), meta["page"])
	assert.Equal(t, float64(20), meta["limit"])
	assert.Equal(t, "failed", meta["status"])
}

func TestAuditHandler_GetLogs_InvalidDateRange(t *testing.T) {
	handler := audit.NewHandler(audit.NewService(&stubAuditRepo{}), nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodGet,
		"/audit-logs?date_from=2026-04-21T00:00:00Z&date_to=2026-04-20T00:00:00Z",
		nil,
	)

	handler.GetLogs(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "error", bodyMap["status"])
	assert.Equal(t, "date_from must be before or equal to date_to", bodyMap["message"])
}

func TestAuditHandler_ExportLogs_ReturnsCSV(t *testing.T) {
	expectedActorID := uint(7)
	service := audit.NewService(&stubAuditRepo{
		findForExport: func(filter audit.Filter) ([]audit.AuditLog, error) {
			assert.NotNil(t, filter.ActorUserID)
			assert.Equal(t, expectedActorID, *filter.ActorUserID)
			return []audit.AuditLog{
				{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC),
					},
					ActorUserID: &expectedActorID,
					Action:      "login",
					Resource:    "auth",
					Status:      "success",
					Description: "user logged in",
					IPAddress:   "127.0.0.1",
					UserAgent:   "Postman",
				},
			}, nil
		},
	})
	handler := audit.NewHandler(service, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/audit-logs/export?actor_user_id=7", nil)

	handler.ExportLogs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment; filename=")
	assert.Contains(t, w.Body.String(), "id,created_at,actor_user_id,action,resource,resource_id,status,description,ip_address,user_agent")
	assert.Contains(t, w.Body.String(), "login,auth")
}

func TestAuditService_ExportLogsCSV_PropagatesRepositoryError(t *testing.T) {
	service := audit.NewService(&stubAuditRepo{
		findForExport: func(filter audit.Filter) ([]audit.AuditLog, error) {
			return nil, errors.New("db down")
		},
	})

	payload, err := service.ExportLogsCSV(audit.Filter{})

	assert.Nil(t, payload)
	assert.EqualError(t, err, "db down")
}

func TestAuditInvestigatorService_WithMockProvider(t *testing.T) {
	aiService, err := aiModule.NewService(config.AIConfig{
		Enabled:  true,
		Provider: "mock",
		Model:    "mock-model",
	})
	assert.NoError(t, err)

	repo := &stubAuditRepo{
		findAllWithFilter: func(filter audit.Filter) ([]audit.AuditLog, int64, error) {
			return []audit.AuditLog{
				{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
					},
					Action:      "login",
					Resource:    "auth",
					Status:      "failed",
					Description: "invalid credentials",
					IPAddress:   "203.0.113.1",
				},
			}, 1, nil
		},
	}

	service := audit.NewInvestigatorService(repo, aiService, audit.NewService(repo))
	result, logs, err := service.Investigate(t.Context(), audit.Filter{Resource: "auth", Limit: 20})

	assert.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Contains(t, result.Summary, "AI mock analysis completed")
	assert.NotEmpty(t, result.Timeline)
	assert.NotEmpty(t, result.Recommendations)
}

func TestAuditHandler_InvestigateLogs_Success(t *testing.T) {
	aiService, err := aiModule.NewService(config.AIConfig{
		Enabled:  true,
		Provider: "mock",
		Model:    "mock-model",
	})
	assert.NoError(t, err)

	repo := &stubAuditRepo{
		findAllWithFilter: func(filter audit.Filter) ([]audit.AuditLog, int64, error) {
			assert.Equal(t, "auth", filter.Resource)
			assert.Equal(t, "failed", filter.Status)
			assert.Equal(t, 25, filter.Limit)
			return []audit.AuditLog{
				{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC),
					},
					Action:      "login",
					Resource:    "auth",
					Status:      "failed",
					Description: "invalid credentials",
					IPAddress:   "203.0.113.1",
				},
			}, 1, nil
		},
		createInvestigation: func(record *audit.AuditInvestigation) error {
			record.ID = 77
			assert.Equal(t, uint(9), *record.CreatedByUserID)
			assert.Equal(t, "mock", record.AIProvider)
			assert.Equal(t, "mock-model", record.AIModel)
			assert.Equal(t, 1, record.LogCount)
			assert.NotEmpty(t, record.SnapshotHash)
			return nil
		},
		createLog: func(log *audit.AuditLog) error {
			assert.Equal(t, "audit_investigation", log.Resource)
			assert.Equal(t, "created", log.Action)
			assert.Equal(t, "success", log.Status)
			return nil
		},
	}

	handler := audit.NewHandler(
		audit.NewService(repo),
		audit.NewInvestigatorService(repo, aiService, audit.NewService(repo)),
	)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/audit-logs/investigate",
		strings.NewReader(`{"resource":"auth","status":"failed","limit":25}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", uint(9))

	handler.InvestigateLogs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "success", bodyMap["status"])
	assert.Equal(t, "Audit investigation completed", bodyMap["message"])
	meta := bodyMap["meta"].(map[string]interface{})
	assert.Equal(t, float64(1), meta["log_count"])
	assert.Equal(t, float64(77), meta["investigation_id"])
}

func TestAuditHandler_InvestigateLogs_DisabledAI(t *testing.T) {
	repo := &stubAuditRepo{}
	handler := audit.NewHandler(
		audit.NewService(repo),
		audit.NewInvestigatorService(repo, nil, audit.NewService(repo)),
	)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/audit-logs/investigate",
		strings.NewReader(`{"resource":"auth"}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.InvestigateLogs(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "error", bodyMap["status"])
	assert.Equal(t, "ai investigator is not enabled", bodyMap["message"])
}

func TestAuditHandler_InvestigateLogs_ReusesExistingInvestigation(t *testing.T) {
	aiService, err := aiModule.NewService(config.AIConfig{
		Enabled:        true,
		Provider:       "mock",
		Model:          "mock-model",
		TimeoutSeconds: 30,
	})
	assert.NoError(t, err)

	repo := &stubAuditRepo{
		findAllWithFilter: func(filter audit.Filter) ([]audit.AuditLog, int64, error) {
			return []audit.AuditLog{
				{
					Model:       gorm.Model{ID: 10, CreatedAt: time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)},
					Action:      "login",
					Resource:    "auth",
					Status:      "failed",
					Description: "invalid credentials",
					IPAddress:   "203.0.113.1",
				},
			}, 1, nil
		},
		findBySnapshot: func(createdByUserID *uint, snapshotHash string) (*audit.AuditInvestigation, error) {
			assert.NotNil(t, createdByUserID)
			assert.Equal(t, uint(9), *createdByUserID)
			assert.NotEmpty(t, snapshotHash)
			return &audit.AuditInvestigation{
				Model:                 gorm.Model{ID: 88, CreatedAt: time.Date(2026, 4, 22, 4, 0, 0, 0, time.UTC)},
				CreatedByUserID:       createdByUserID,
				SnapshotHash:          snapshotHash,
				Resource:              "auth",
				Status:                "failed",
				LogCount:              1,
				Summary:               "existing summary",
				TimelineJSON:          `["item 1"]`,
				SuspiciousSignalsJSON: `["signal 1"]`,
				RecommendationsJSON:   `["recommendation 1"]`,
			}, nil
		},
		createLog: func(log *audit.AuditLog) error {
			assert.Equal(t, "reused", log.Action)
			assert.Equal(t, "audit_investigation", log.Resource)
			return nil
		},
		createInvestigation: func(record *audit.AuditInvestigation) error {
			t.Fatal("did not expect a new investigation to be created")
			return nil
		},
	}

	handler := audit.NewHandler(
		audit.NewService(repo),
		audit.NewInvestigatorService(repo, aiService, audit.NewService(repo)),
	)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/audit-logs/investigate",
		strings.NewReader(`{"resource":"auth","status":"failed","limit":25}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", uint(9))

	handler.InvestigateLogs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	bodyMap := decodeBodyMap(t, w)
	meta := bodyMap["meta"].(map[string]interface{})
	assert.Equal(t, float64(88), meta["investigation_id"])
	assert.Equal(t, true, meta["reused_existing"])
}

func TestAuditHandler_ListInvestigations_Success(t *testing.T) {
	repo := &stubAuditRepo{
		findInvestigations: func(filter audit.InvestigationFilter) ([]audit.AuditInvestigation, int64, error) {
			assert.Equal(t, 2, filter.Page)
			assert.Equal(t, 5, filter.Limit)
			assert.Equal(t, "auth", filter.Resource)
			assert.Equal(t, "failed", filter.Status)
			assert.Equal(t, "ollama", filter.AIProvider)
			assert.Equal(t, "qwen2.5:3b", filter.AIModel)
			assert.Equal(t, "invalid credentials", filter.Search)
			assert.NotNil(t, filter.CreatedByUserID)
			assert.Equal(t, uint(9), *filter.CreatedByUserID)
			assert.NotNil(t, filter.CreatedFrom)
			assert.NotNil(t, filter.CreatedTo)
			return []audit.AuditInvestigation{
				{
					Model:                 gorm.Model{ID: 11, CreatedAt: time.Date(2026, 4, 22, 4, 0, 0, 0, time.UTC)},
					Resource:              "auth",
					Status:                "failed",
					Summary:               "Repeated failed login attempts detected.",
					TimelineJSON:          `["item 1"]`,
					SuspiciousSignalsJSON: `["signal 1"]`,
					RecommendationsJSON:   `["recommendation 1"]`,
				},
			}, 1, nil
		},
	}

	handler := audit.NewHandler(
		audit.NewService(repo),
		audit.NewInvestigatorService(repo, &aiModule.Service{}, audit.NewService(repo)),
	)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/audit-logs/investigations?page=2&limit=5&resource=auth&status=failed&created_by_user_id=9&ai_provider=ollama&ai_model=qwen2.5:3b&search=invalid%20credentials&created_from=2026-04-20T00:00:00Z&created_to=2026-04-22T23:59:59Z", nil)

	handler.ListInvestigations(c)

	assert.Equal(t, http.StatusOK, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "success", bodyMap["status"])
	assert.Equal(t, "Audit investigations fetched", bodyMap["message"])
	meta := bodyMap["meta"].(map[string]interface{})
	assert.Equal(t, float64(2), meta["page"])
	assert.Equal(t, float64(5), meta["limit"])
	assert.Equal(t, "auth", meta["resource"])
}

func TestAuditHandler_ListInvestigations_InvalidDateRange(t *testing.T) {
	handler := audit.NewHandler(
		audit.NewService(&stubAuditRepo{}),
		audit.NewInvestigatorService(&stubAuditRepo{}, &aiModule.Service{}, audit.NewService(&stubAuditRepo{})),
	)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/audit-logs/investigations?created_from=2026-04-22T00:00:00Z&created_to=2026-04-20T00:00:00Z", nil)

	handler.ListInvestigations(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "error", bodyMap["status"])
	assert.Equal(t, "created_from must be before or equal to created_to", bodyMap["message"])
}

func TestAuditHandler_GetInvestigationByID_Success(t *testing.T) {
	repo := &stubAuditRepo{
		findInvestigationByID: func(id uint) (*audit.AuditInvestigation, error) {
			assert.Equal(t, uint(11), id)
			return &audit.AuditInvestigation{
				Model:                 gorm.Model{ID: 11, CreatedAt: time.Date(2026, 4, 22, 4, 0, 0, 0, time.UTC)},
				Resource:              "auth",
				Status:                "failed",
				Summary:               "Repeated failed login attempts detected.",
				TimelineJSON:          `["item 1"]`,
				SuspiciousSignalsJSON: `["signal 1"]`,
				RecommendationsJSON:   `["recommendation 1"]`,
			}, nil
		},
	}

	handler := audit.NewHandler(
		audit.NewService(repo),
		audit.NewInvestigatorService(repo, &aiModule.Service{}, audit.NewService(repo)),
	)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "11"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/audit-logs/investigations/11", nil)

	handler.GetInvestigationByID(c)

	assert.Equal(t, http.StatusOK, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "success", bodyMap["status"])
	assert.Equal(t, "Audit investigation fetched", bodyMap["message"])
}

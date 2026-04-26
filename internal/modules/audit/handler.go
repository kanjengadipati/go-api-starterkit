package audit

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"pleco-api/internal/ai"
	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	AuditService *Service
	AIService    *InvestigatorService
}

func NewHandler(auditService *Service, aiService *InvestigatorService) *Handler {
	return &Handler{AuditService: auditService, AIService: aiService}
}

func (h *Handler) GetLogs(c *gin.Context) {
	filter, err := buildFilter(c)
	if err != nil {
		httpx.Error(c, 400, err.Error())
		return
	}

	logs, total, err := h.AuditService.GetLogs(filter)
	if err != nil {
		httpx.Error(c, 500, "Failed to fetch audit logs")
		return
	}

	httpx.Success(c, 200, "Audit logs fetched", logs, gin.H{
		"page":          filter.Page,
		"limit":         filter.Limit,
		"total":         total,
		"action":        filter.Action,
		"resource":      filter.Resource,
		"status":        filter.Status,
		"actor_user_id": filter.ActorUserID,
		"search":        filter.Search,
		"date_from":     formatTime(filter.DateFrom),
		"date_to":       formatTime(filter.DateTo),
	})
}

func (h *Handler) ExportLogs(c *gin.Context) {
	filter, err := buildFilter(c)
	if err != nil {
		httpx.Error(c, 400, err.Error())
		return
	}

	payload, err := h.AuditService.ExportLogsCSV(filter)
	if err != nil {
		httpx.Error(c, 500, "Failed to export audit logs")
		return
	}

	filename := fmt.Sprintf("audit-logs-%s.csv", time.Now().UTC().Format("20060102-150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(200, "text/csv; charset=utf-8", payload)
}

func (h *Handler) InvestigateLogs(c *gin.Context) {
	var input InvestigateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	filter, err := buildFilterFromRequest(input)
	if err != nil {
		httpx.Error(c, 400, err.Error())
		return
	}

	result, logs, err := h.AIService.Investigate(c.Request.Context(), filter)
	if err != nil {
		switch {
		case err.Error() == "ai investigator is not enabled":
			httpx.Error(c, 503, err.Error())
		case err.Error() == "no audit logs found for investigation":
			httpx.Error(c, 404, err.Error())
		case errors.Is(err, ai.ErrTimeout):
			httpx.Error(c, 504, "ai investigation timed out")
		case errors.Is(err, ai.ErrInvalidStructuredOutput):
			httpx.Error(c, 502, "ai returned an invalid structured response")
		default:
			httpx.Error(c, 500, err.Error())
		}
		return
	}

	var investigationID uint
	reusedExisting := false
	if saved, reused, err := h.AIService.SaveInvestigation(currentUserID(c), filter, result, logs, c.ClientIP(), c.Request.UserAgent()); err != nil {
		httpx.Error(c, 500, "Failed to save audit investigation")
		return
	} else if saved != nil {
		investigationID = saved.ID
		reusedExisting = reused
	}

	httpx.Success(c, 200, "Audit investigation completed", result, gin.H{
		"investigation_id": investigationID,
		"reused_existing":  reusedExisting,
		"log_count":        len(logs),
		"limit":            filter.Limit,
		"resource":         filter.Resource,
		"action":           filter.Action,
		"status":           filter.Status,
	})
}

func (h *Handler) ListInvestigations(c *gin.Context) {
	filter, err := buildInvestigationFilter(c)
	if err != nil {
		httpx.Error(c, 400, err.Error())
		return
	}

	items, total, err := h.AIService.ListInvestigations(filter)
	if err != nil {
		httpx.Error(c, 500, "Failed to fetch audit investigations")
		return
	}

	httpx.Success(c, 200, "Audit investigations fetched", items, gin.H{
		"page":               filter.Page,
		"limit":              filter.Limit,
		"total":              total,
		"resource":           filter.Resource,
		"status":             filter.Status,
		"created_by_user_id": filter.CreatedByUserID,
		"ai_provider":        filter.AIProvider,
		"ai_model":           filter.AIModel,
		"search":             filter.Search,
		"created_from":       formatTime(filter.CreatedFrom),
		"created_to":         formatTime(filter.CreatedTo),
	})
}

func (h *Handler) GetInvestigationByID(c *gin.Context) {
	id, err := parsePositiveUintParam(c.Param("id"), "id")
	if err != nil {
		httpx.Error(c, 400, err.Error())
		return
	}

	item, err := h.AIService.GetInvestigationByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpx.Error(c, 404, "audit investigation not found")
			return
		}
		httpx.Error(c, 500, "Failed to fetch audit investigation")
		return
	}

	httpx.Success(c, 200, "Audit investigation fetched", item, nil)
}

func buildFilter(c *gin.Context) (Filter, error) {
	page, limit := paginationFromQuery(c)

	filter := Filter{
		Page:     page,
		Limit:    limit,
		Action:   c.Query("action"),
		Resource: c.Query("resource"),
		Status:   c.Query("status"),
		Search:   c.Query("search"),
	}

	if actorID := c.Query("actor_user_id"); actorID != "" {
		parsed, err := strconv.ParseUint(actorID, 10, 64)
		if err != nil || parsed == 0 {
			return Filter{}, fmt.Errorf("actor_user_id must be a positive integer")
		}
		value := uint(parsed)
		filter.ActorUserID = &value
	}

	var err error
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		parsed := time.Time{}
		parsed, err = time.Parse(time.RFC3339, dateFrom)
		if err != nil {
			return Filter{}, fmt.Errorf("date_from must use RFC3339 format")
		}
		filter.DateFrom = &parsed
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		parsed := time.Time{}
		parsed, err = time.Parse(time.RFC3339, dateTo)
		if err != nil {
			return Filter{}, fmt.Errorf("date_to must use RFC3339 format")
		}
		filter.DateTo = &parsed
	}
	if filter.DateFrom != nil && filter.DateTo != nil && filter.DateFrom.After(*filter.DateTo) {
		return Filter{}, fmt.Errorf("date_from must be before or equal to date_to")
	}

	return filter, nil
}

func paginationFromQuery(c *gin.Context) (int, int) {
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}

	return page, limit
}

func buildInvestigationFilter(c *gin.Context) (InvestigationFilter, error) {
	page, limit := paginationFromQuery(c)

	filter := InvestigationFilter{
		Page:       page,
		Limit:      limit,
		Resource:   c.Query("resource"),
		Status:     c.Query("status"),
		AIProvider: c.Query("ai_provider"),
		AIModel:    c.Query("ai_model"),
		Search:     c.Query("search"),
	}

	if createdBy := c.Query("created_by_user_id"); createdBy != "" {
		value, err := parsePositiveUintParam(createdBy, "created_by_user_id")
		if err != nil {
			return InvestigationFilter{}, err
		}
		filter.CreatedByUserID = &value
	}

	var err error
	if createdFrom := c.Query("created_from"); createdFrom != "" {
		parsed := time.Time{}
		parsed, err = time.Parse(time.RFC3339, createdFrom)
		if err != nil {
			return InvestigationFilter{}, fmt.Errorf("created_from must use RFC3339 format")
		}
		filter.CreatedFrom = &parsed
	}
	if createdTo := c.Query("created_to"); createdTo != "" {
		parsed := time.Time{}
		parsed, err = time.Parse(time.RFC3339, createdTo)
		if err != nil {
			return InvestigationFilter{}, fmt.Errorf("created_to must use RFC3339 format")
		}
		filter.CreatedTo = &parsed
	}
	if filter.CreatedFrom != nil && filter.CreatedTo != nil && filter.CreatedFrom.After(*filter.CreatedTo) {
		return InvestigationFilter{}, fmt.Errorf("created_from must be before or equal to created_to")
	}

	return filter, nil
}

func parsePositiveUintParam(raw string, field string) (uint, error) {
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		return 0, fmt.Errorf("%s must be a positive integer", field)
	}
	return uint(value), nil
}

func buildFilterFromRequest(input InvestigateRequest) (Filter, error) {
	filter := Filter{
		Page:     1,
		Limit:    input.Limit,
		Action:   input.Action,
		Resource: input.Resource,
		Status:   input.Status,
		Search:   input.Search,
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	filter.ActorUserID = input.ActorUserID

	var err error
	if input.DateFrom != "" {
		parsed := time.Time{}
		parsed, err = time.Parse(time.RFC3339, input.DateFrom)
		if err != nil {
			return Filter{}, fmt.Errorf("date_from must use RFC3339 format")
		}
		filter.DateFrom = &parsed
	}
	if input.DateTo != "" {
		parsed := time.Time{}
		parsed, err = time.Parse(time.RFC3339, input.DateTo)
		if err != nil {
			return Filter{}, fmt.Errorf("date_to must use RFC3339 format")
		}
		filter.DateTo = &parsed
	}
	if filter.DateFrom != nil && filter.DateTo != nil && filter.DateFrom.After(*filter.DateTo) {
		return Filter{}, fmt.Errorf("date_from must be before or equal to date_to")
	}

	return filter, nil
}

func formatTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}

func currentUserID(c *gin.Context) *uint {
	raw, ok := c.Get("user_id")
	if !ok {
		return nil
	}
	value, ok := raw.(uint)
	if !ok {
		return nil
	}
	return &value
}

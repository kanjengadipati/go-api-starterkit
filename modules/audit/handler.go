package audit

import (
	"strconv"

	"go-api-starterkit/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	AuditService *Service
}

func NewHandler(auditService *Service) *Handler {
	return &Handler{AuditService: auditService}
}

func (h *Handler) GetLogs(c *gin.Context) {
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
		}
	}

	action := c.Query("action")
	resource := c.Query("resource")

	logs, total, err := h.AuditService.GetLogs(page, limit, action, resource)
	if err != nil {
		httpx.Error(c, 500, "Failed to fetch audit logs")
		return
	}

	httpx.Success(c, 200, "Audit logs fetched", logs, gin.H{
		"page":     page,
		"limit":    limit,
		"total":    total,
		"action":   action,
		"resource": resource,
	})
}

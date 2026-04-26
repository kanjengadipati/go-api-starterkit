package audit

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/modules/permission"
	"pleco-api/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permissionService *permission.Service, tokenVersionSrc middleware.AccessTokenVersionSource) {
	protected := api.Group("/auth")
	protected.Use(middleware.AuthMiddleware(jwtService))

	admin := protected.Group("/admin")
	admin.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))
	admin.GET("/audit-logs", middleware.RequirePermission(permissionService, "audit.read"), handler.GetLogs)
	admin.GET("/audit-logs/export", middleware.RequirePermission(permissionService, "audit.read"), handler.ExportLogs)
	admin.POST("/audit-logs/investigate", middleware.RequirePermission(permissionService, "audit.investigate"), handler.InvestigateLogs)
	admin.GET("/audit-logs/investigations", middleware.RequirePermission(permissionService, "audit.read"), handler.ListInvestigations)
	admin.GET("/audit-logs/investigations/:id", middleware.RequirePermission(permissionService, "audit.read"), handler.GetInvestigationByID)
}

package audit

import (
	"go-api-starterkit/middleware"
	"go-api-starterkit/modules/permission"
	"go-api-starterkit/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permissionService *permission.Service) {
	protected := api.Group("/auth")
	protected.Use(middleware.AuthMiddleware(jwtService))

	admin := protected.Group("/admin")
	admin.GET("/audit-logs", middleware.RequirePermission(permissionService, "audit.read"), handler.GetLogs)
}

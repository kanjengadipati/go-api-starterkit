package role

import (
	"go-api-starterkit/internal/middleware"
	permissionModule "go-api-starterkit/internal/modules/permission"
	"go-api-starterkit/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permissionService *permissionModule.Service, tokenVersionSrc middleware.AccessTokenVersionSource) {
	protected := api.Group("/auth/admin")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.Use(middleware.RequireAccessTokenVersion(tokenVersionSrc))

	protected.GET("/roles", middleware.RequirePermission(permissionService, "role.read"), handler.GetRoles)
	protected.GET("/permissions", middleware.RequirePermission(permissionService, "permission.read"), handler.GetPermissions)
	protected.GET("/roles/:id/permissions", middleware.RequirePermission(permissionService, "role.read"), handler.GetRolePermissions)
	protected.PUT("/roles/:id/permissions", middleware.RequirePermission(permissionService, "role.update_permissions"), handler.UpdateRolePermissions)
}

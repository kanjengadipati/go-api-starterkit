package user

import (
	"go-api-starterkit/middleware"
	"go-api-starterkit/modules/permission"
	"go-api-starterkit/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *Handler, jwtService *services.JWTService, permissionService *permission.Service) {
	protected := api.Group("/auth")
	protected.Use(middleware.AuthMiddleware(jwtService))

	protected.PATCH("/profile", handler.UpdateProfile)
	protected.PATCH("/change-password", handler.ChangePassword)

	admin := protected.Group("/admin")
	admin.GET("/users", middleware.RequirePermission(permissionService, "user.read_all"), handler.GetAllUsers)
	admin.GET("/users/:id", middleware.RequirePermission(permissionService, "user.read"), handler.GetUserByID)
	admin.POST("/users", middleware.RequirePermission(permissionService, "user.create"), handler.CreateUser)
	admin.PUT("/users/:id", middleware.RequirePermission(permissionService, "user.update"), handler.UpdateUser)
	admin.DELETE("/users/:id", middleware.RequirePermission(permissionService, "user.delete"), handler.DeleteUser)
}

package controllers

import (
	"fmt"
	"go-auth-app/dto"
	"go-auth-app/repositories"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	UserRepo repositories.UserRepository
}

func (u *UserController) GetAllUsers(c *gin.Context) {
	// default values
	page := 1
	limit := 10

	// query params
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	search := c.Query("search")
	role := c.Query("role")

	users, total, err := u.UserRepo.FindAllWithFilter(page, limit, search, role)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch users"})
		return
	}

	var result []dto.UserResponse

	for _, user := range users {
		result = append(result, dto.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		})
	}

	c.JSON(200, gin.H{
		"data": result,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (u *UserController) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")

	var id uint
	fmt.Sscanf(idParam, "%d", &id)

	err := u.UserRepo.Delete(id)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed"})
		return
	}

	c.JSON(200, gin.H{"message": "Deleted"})
}

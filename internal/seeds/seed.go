package seeds

import (
	"fmt"
	"go-api-starterkit/internal/config"
	permissionModule "go-api-starterkit/internal/modules/permission"
	roleModule "go-api-starterkit/internal/modules/role"
	userModule "go-api-starterkit/internal/modules/user"
	"go-api-starterkit/internal/services"
	"log"

	"gorm.io/gorm"
)

// Helper to check if db is nil and log fatal if so
func mustHaveDB(db *gorm.DB) {
	if db == nil {
		log.Fatal("DB connection failed: db is nil")
	}
}

func SeedRoles(db *gorm.DB) map[string]roleModule.Role {
	mustHaveDB(db)

	roleNames := []string{"superadmin", "admin", "user"}
	roleMap := make(map[string]roleModule.Role)

	for _, name := range roleNames {
		role := roleModule.Role{Name: name}
		if err := db.FirstOrCreate(&role, roleModule.Role{Name: name}).Error; err != nil {
			log.Printf("Failed to seed role %s: %v", name, err)
			continue
		}
		roleMap[name] = role
	}
	fmt.Println("Roles seeding done")
	return roleMap
}

func SeedPermissions(db *gorm.DB) map[string]permissionModule.Permission {
	mustHaveDB(db)

	permNames := []string{
		"dashboard.view",
		"user.read_all",
		"user.read",
		"user.create",
		"user.update",
		"user.delete",
		"audit.read",
		"audit.investigate",
		"role.read",
		"permission.read",
		"role.update_permissions",
		"session.read",
		"session.delete",
	}
	permMap := make(map[string]permissionModule.Permission)

	for _, name := range permNames {
		perm := permissionModule.Permission{Name: name}
		if err := db.FirstOrCreate(&perm, permissionModule.Permission{Name: name}).Error; err != nil {
			log.Printf("Failed to seed permission %s: %v", name, err)
			continue
		}
		permMap[name] = perm
	}
	fmt.Println("Permissions seeding done")
	return permMap
}

func SeedRolePermissions(db *gorm.DB) {
	mustHaveDB(db)

	roleMap := SeedRoles(db)

	rolePermissions := map[string][]string{
		"admin": {
			"dashboard.view",
			"user.read_all",
			"user.read",
			"user.create",
			"user.update",
			"user.delete",
			"audit.read",
			"audit.investigate",
			"role.read",
			"permission.read",
			"role.update_permissions",
			"session.read",
			"session.delete",
		},
		"user": {
			"dashboard.view",
			"session.read",
		},
	}

	for roleName, permissions := range rolePermissions {
		role, ok := roleMap[roleName]
		if !ok {
			log.Printf("Role %s not found while seeding role permissions", roleName)
			continue
		}

		for _, permission := range permissions {
			type rolePermissionRow struct {
				ID         uint   `gorm:"column:id"`
				RoleID     uint   `gorm:"column:role_id"`
				Permission string `gorm:"column:permission"`
			}

			row := rolePermissionRow{RoleID: role.ID, Permission: permission}
			if err := db.Table("role_permissions").
				Where("role_id = ? AND permission = ?", role.ID, permission).
				FirstOrCreate(&row).Error; err != nil {
				log.Printf("Failed to seed role permission %s -> %s: %v", roleName, permission, err)
			}
		}
	}

	fmt.Println("Role permissions seeding done")
}

func SeedAdmin(db *gorm.DB, cfg config.AppConfig) {
	mustHaveDB(db)

	var user userModule.User

	roleMap := SeedRoles(db)
	superadminRole, ok := roleMap["superadmin"]
	if !ok {
		log.Println("Superadmin role not found: cannot seed admin user")
		return
	}

	email := cfg.AdminEmail
	password := cfg.AdminPassword

	if email == "" || password == "" {
		log.Println("ADMIN_EMAIL or ADMIN_PASSWORD environment variables not set, skipping admin seeding")
		return
	}

	if err := db.Where("email = ?", email).First(&user).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Error checking for admin user: %v", err)
		return
	}
	if user.ID == 0 {
		hashedPassword, err := services.HashPassword(password)
		if err != nil {
			log.Printf("Error hashing admin password: %v", err)
			return
		}

		admin := userModule.User{
			Name:       "Super Admin",
			Email:      email,
			Password:   hashedPassword,
			RoleID:     superadminRole.ID,
			Role:       superadminRole.Name,
			IsVerified: true,
		}

		if err := db.Create(&admin).Error; err != nil {
			log.Printf("Error creating admin user: %v", err)
			return
		} else {
			log.Printf("Super Admin user seeded with email: %s", email)
		}
	}
	fmt.Println("Admin seeding done")
}

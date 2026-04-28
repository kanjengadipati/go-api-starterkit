package permission

import "gorm.io/gorm"

type Repository interface {
	HasRolePermission(roleName, permission string) (bool, error)
	ListAllPermissions() ([]Permission, error)
	ListRolePermissions(roleID uint) ([]string, error)
	ListRolePermissionsByName(roleName string) ([]string, error)
	AllPermissionsExist(names []string) (bool, error)
	ReplaceRolePermissions(roleID uint, permissions []string) error
	WithTx(tx *gorm.DB) Repository
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) HasRolePermission(roleName, permission string) (bool, error) {
	var count int64

	err := r.db.Table("role_permissions").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Where("roles.name = ? AND role_permissions.permission = ?", roleName, permission).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *gormRepository) ListAllPermissions() ([]Permission, error) {
	var permissions []Permission
	if err := r.db.Order("name ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *gormRepository) ListRolePermissions(roleID uint) ([]string, error) {
	var permissions []string
	if err := r.db.Table("role_permissions").
		Where("role_id = ?", roleID).
		Order("permission ASC").
		Pluck("permission", &permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *gormRepository) ListRolePermissionsByName(roleName string) ([]string, error) {
	var permissions []string
	if err := r.db.Table("role_permissions").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Where("roles.name = ?", roleName).
		Order("role_permissions.permission ASC").
		Pluck("role_permissions.permission", &permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *gormRepository) AllPermissionsExist(names []string) (bool, error) {
	if len(names) == 0 {
		return true, nil
	}

	var count int64
	if err := r.db.Model(&Permission{}).Where("name IN ?", names).Count(&count).Error; err != nil {
		return false, err
	}

	return count == int64(len(names)), nil
}

func (r *gormRepository) ReplaceRolePermissions(roleID uint, permissions []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("role_permissions").Where("role_id = ?", roleID).Delete(map[string]interface{}{}).Error; err != nil {
			return err
		}

		for _, permission := range permissions {
			values := map[string]interface{}{
				"role_id":    roleID,
				"permission": permission,
			}
			if err := tx.Table("role_permissions").Create(values).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *gormRepository) WithTx(tx *gorm.DB) Repository {
	return &gormRepository{db: tx}
}

package role

import (
	"errors"
	"slices"

	permissionModule "pleco-api/internal/modules/permission"
)

type Service struct {
	RoleRepo       Repository
	PermissionRepo permissionModule.Repository
}

func NewService(roleRepo Repository, permissionRepo permissionModule.Repository) *Service {
	return &Service{
		RoleRepo:       roleRepo,
		PermissionRepo: permissionRepo,
	}
}

func (s *Service) FindByID(id uint) (*Role, error) {
	return s.RoleRepo.FindByID(id)
}

func (s *Service) ListRoles() ([]Role, error) {
	return s.RoleRepo.FindAll()
}

func (s *Service) ListPermissions() ([]permissionModule.Permission, error) {
	return s.PermissionRepo.ListAllPermissions()
}

func (s *Service) GetRolePermissions(roleID uint) (*Role, []string, error) {
	role, err := s.RoleRepo.FindByID(roleID)
	if err != nil {
		return nil, nil, err
	}

	permissions, err := s.PermissionRepo.ListRolePermissions(roleID)
	if err != nil {
		return nil, nil, err
	}

	return role, permissions, nil
}

func (s *Service) UpdateRolePermissions(roleID uint, permissions []string) (*Role, []string, error) {
	role, err := s.RoleRepo.FindByID(roleID)
	if err != nil {
		return nil, nil, err
	}

	normalized := normalizePermissions(permissions)
	exists, err := s.PermissionRepo.AllPermissionsExist(normalized)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return nil, nil, errors.New("one or more permissions are invalid")
	}

	if err := s.PermissionRepo.ReplaceRolePermissions(roleID, normalized); err != nil {
		return nil, nil, err
	}

	return role, normalized, nil
}

func normalizePermissions(permissions []string) []string {
	unique := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		if permission == "" {
			continue
		}
		if slices.Contains(unique, permission) {
			continue
		}
		unique = append(unique, permission)
	}

	slices.Sort(unique)
	return unique
}

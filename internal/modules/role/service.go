package role

import (
	"context"
	"fmt"
	"slices"
	"time"

	"pleco-api/internal/cache"
	"pleco-api/internal/domain"
	"pleco-api/internal/modules/audit"
	permissionModule "pleco-api/internal/modules/permission"
)

type Service struct {
	RoleRepo       Repository
	PermissionRepo permissionModule.Repository
	Cache          cache.Store
	AuditSvc       *audit.Service
}

func NewService(roleRepo Repository, permissionRepo permissionModule.Repository) *Service {
	return &Service{
		RoleRepo:       roleRepo,
		PermissionRepo: permissionRepo,
	}
}

func (s *Service) FindByID(id uint) (*Role, error) {
	if s.Cache != nil {
		var role Role
		key := fmt.Sprintf("role:%d", id)
		if ok, err := s.Cache.GetJSON(context.Background(), key, &role); err == nil && ok {
			return &role, nil
		}
		found, err := s.RoleRepo.FindByID(id)
		if err != nil {
			return nil, err
		}
		_ = s.Cache.SetJSON(context.Background(), key, found, 15*time.Minute)
		return found, nil
	}

	return s.RoleRepo.FindByID(id)
}

func (s *Service) ListRoles() ([]Role, error) {
	if s.Cache != nil {
		var roles []Role
		if ok, err := s.Cache.GetJSON(context.Background(), "roles", &roles); err == nil && ok {
			return roles, nil
		}
		roles, err := s.RoleRepo.FindAll()
		if err != nil {
			return nil, err
		}
		_ = s.Cache.SetJSON(context.Background(), "roles", roles, 20*time.Minute)
		return roles, nil
	}

	return s.RoleRepo.FindAll()
}

func (s *Service) ListPermissions() ([]permissionModule.Permission, error) {
	return s.PermissionRepo.ListAllPermissions()
}

func (s *Service) GetRolePermissions(roleID uint) (*Role, []string, error) {
	if s.Cache != nil {
		var cached RolePermissionsResponse
		key := fmt.Sprintf("role:%d:permissions", roleID)
		if ok, err := s.Cache.GetJSON(context.Background(), key, &cached); err == nil && ok {
			return &Role{ID: cached.ID, Name: cached.Name}, cached.Permissions, nil
		}

		role, permissions, err := s.getRolePermissionsFromRepo(roleID)
		if err != nil {
			return nil, nil, err
		}
		_ = s.Cache.SetJSON(context.Background(), key, RolePermissionsResponse{
			ID:          role.ID,
			Name:        role.Name,
			Permissions: permissions,
		}, 15*time.Minute)
		return role, permissions, nil
	}

	return s.getRolePermissionsFromRepo(roleID)
}

func (s *Service) getRolePermissionsFromRepo(roleID uint) (*Role, []string, error) {
	role, err := s.FindByID(roleID)
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
	role, err := s.FindByID(roleID)
	if err != nil {
		return nil, nil, err
	}

	normalized := normalizePermissions(permissions)
	exists, err := s.PermissionRepo.AllPermissionsExist(normalized)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return nil, nil, domain.ErrInvalidPermission
	}

	if err := s.PermissionRepo.ReplaceRolePermissions(roleID, normalized); err != nil {
		return nil, nil, err
	}
	if s.Cache != nil {
		_ = s.Cache.Delete(context.Background(), fmt.Sprintf("role:%d:permissions", roleID), fmt.Sprintf("role:%d", roleID), "roles")
		_ = s.Cache.DeletePrefix(context.Background(), fmt.Sprintf("role:permission:%s:", role.Name))
		_ = s.Cache.Delete(context.Background(), fmt.Sprintf("role:permissions:%s", role.Name))
		_ = s.Cache.DeletePrefix(context.Background(), "user:permissions:")
		_ = s.Cache.DeletePrefix(context.Background(), "user:profile:")
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

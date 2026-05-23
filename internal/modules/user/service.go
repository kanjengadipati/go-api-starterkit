package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pleco-api/internal/cache"
	"pleco-api/internal/modules/audit"
	tokenModule "pleco-api/internal/modules/token"
	"pleco-api/internal/services"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var errUnsupportedRole = errors.New("unsupported role")

type Service struct {
	DB               *gorm.DB
	UserRepo         Repository
	RefreshTokenRepo tokenModule.RefreshTokenRepository
	AuditSvc         *audit.Service
	Cache            cache.Store
}

func NewService(db *gorm.DB, userRepo Repository, refreshRepo tokenModule.RefreshTokenRepository, auditSvc *audit.Service) *Service {
	return &Service{DB: db, UserRepo: userRepo, RefreshTokenRepo: refreshRepo, AuditSvc: auditSvc}
}

func (s *Service) GetAllUsers(page, limit int, search, role string) ([]User, int64, error) {
	return s.UserRepo.FindAllWithFilter(page, limit, search, role)
}

func (s *Service) GetUserByID(id uint) (*User, error) {
	if s.Cache != nil {
		var user User
		key := fmt.Sprintf("user:detail:%d", id)
		if ok, err := s.Cache.GetJSON(context.Background(), key, &user); err == nil && ok {
			return &user, nil
		}
		found, err := s.UserRepo.FindByID(id)
		if err != nil {
			return nil, err
		}
		_ = s.Cache.SetJSON(context.Background(), key, found, 5*time.Minute)
		return found, nil
	}

	return s.UserRepo.FindByID(id)
}

func (s *Service) CreateUser(input CreateUserRequest) (*User, error) {
	if existing, err := s.UserRepo.FindByEmail(input.Email); err == nil && existing != nil && existing.ID != 0 {
		return nil, errors.New("email already in use")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	role := input.Role
	if role == "" {
		role = "user"
	}
	if !isAssignableRole(role) {
		return nil, errUnsupportedRole
	}

	hashedPassword, err := services.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &User{
		Name:               input.Name,
		Email:              input.Email,
		Password:           hashedPassword,
		Role:               role,
		IsVerified:         input.IsVerified,
		PasswordUpdatedAt:  now,
		LastPasswordChange: ptrTime(now),
	}

	if err := s.UserRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) UpdateUser(id uint, input UpdateUserRequest) (*User, error) {
	user, err := s.UserRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if !isAssignableRole(input.Role) {
		return nil, errUnsupportedRole
	}

	if existing, err := s.UserRepo.FindByEmail(input.Email); err == nil && existing != nil && existing.ID != 0 && existing.ID != id {
		return nil, errors.New("email already in use")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	oldRole := user.Role
	user.Name = input.Name
	user.Email = input.Email
	user.Role = input.Role
	user.IsVerified = input.IsVerified

	if oldRole != user.Role {
		user.AccessTokenVersion++
	}

	if oldRole != user.Role {
		if err := s.runInTx(func(userRepo Repository, refreshRepo tokenModule.RefreshTokenRepository) error {
			if err := userRepo.Update(user); err != nil {
				return err
			}
			return refreshRepo.DeleteByUser(user.ID)
		}); err != nil {
			return nil, err
		}
		s.invalidateUserCache(user.ID)
		return user, nil
	}

	if err := s.UserRepo.Update(user); err != nil {
		return nil, err
	}
	s.invalidateUserCache(user.ID)
	return user, nil
}

func (s *Service) UpdateProfile(id uint, input UpdateProfileRequest) (*User, error) {
	user, err := s.UserRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	user.Name = input.Name
	user.PhoneNumber = input.PhoneNumber
	if err := s.UserRepo.Update(user); err != nil {
		return nil, err
	}
	s.invalidateUserCache(user.ID)

	return user, nil
}

func (s *Service) ChangePassword(id uint, currentPassword, newPassword string) error {
	user, err := s.UserRepo.FindByID(id)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := services.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	now := time.Now()
	user.PasswordUpdatedAt = now
	user.LastPasswordChange = &now
	user.AccessTokenVersion++

	if err := s.runInTx(func(userRepo Repository, refreshRepo tokenModule.RefreshTokenRepository) error {
		if err := userRepo.Update(user); err != nil {
			return err
		}
		return refreshRepo.DeleteByUser(user.ID)
	}); err != nil {
		return err
	}
	s.invalidateUserCache(user.ID)
	return nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func (s *Service) DeleteUser(id uint, callerRole string, callerID uint) error {
	if id == callerID {
		return errors.New("cannot delete yourself")
	}

	targetUser, err := s.UserRepo.FindByID(id)
	if err != nil {
		return err
	}

	if callerRole == "admin" && targetUser.Role != "user" {
		return errors.New("admin can only delete standard users")
	}

	if targetUser.Role == "superadmin" {
		return errors.New("cannot delete superadmin")
	}

	if err := s.UserRepo.Delete(id); err != nil {
		return err
	}
	s.invalidateUserCache(id)
	return nil
}

func (s *Service) invalidateUserCache(userID uint) {
	if s.Cache == nil {
		return
	}
	_ = s.Cache.Delete(
		context.Background(),
		fmt.Sprintf("user:detail:%d", userID),
		fmt.Sprintf("user:profile:%d", userID),
		fmt.Sprintf("user:permissions:%d", userID),
	)
}

func (s *Service) runInTx(fn func(userRepo Repository, refreshRepo tokenModule.RefreshTokenRepository) error) error {
	if s.DB == nil {
		return fn(s.UserRepo, s.RefreshTokenRepo)
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		return fn(s.UserRepo.WithTx(tx), s.RefreshTokenRepo.WithTx(tx))
	})
}

func isAssignableRole(role string) bool {
	switch role {
	case "admin", "user":
		return true
	default:
		return false
	}
}

package user

import (
	"errors"
	"time"

	"go-api-starterkit/modules/audit"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	UserRepo Repository
	AuditSvc *audit.Service
}

func NewService(userRepo Repository, auditSvc *audit.Service) *Service {
	return &Service{UserRepo: userRepo, AuditSvc: auditSvc}
}

func (s *Service) GetAllUsers(page, limit int, search, role string) ([]User, int64, error) {
	return s.UserRepo.FindAllWithFilter(page, limit, search, role)
}

func (s *Service) GetUserByID(id uint) (*User, error) {
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
	if err != nil {
		return nil, err
	}

	user := &User{
		Name:       input.Name,
		Email:      input.Email,
		Password:   string(hashedPassword),
		Role:       role,
		IsVerified: input.IsVerified,
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

	if existing, err := s.UserRepo.FindByEmail(input.Email); err == nil && existing != nil && existing.ID != 0 && existing.ID != id {
		return nil, errors.New("email already in use")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	user.Name = input.Name
	user.Email = input.Email
	user.Role = input.Role
	user.IsVerified = input.IsVerified

	if err := s.UserRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) UpdateProfile(id uint, input UpdateProfileRequest) (*User, error) {
	user, err := s.UserRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	user.Name = input.Name
	if err := s.UserRepo.Update(user); err != nil {
		return nil, err
	}

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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 14)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.PasswordUpdatedAt = time.Now()

	return s.UserRepo.Update(user)
}

func (s *Service) DeleteUser(id uint) error {
	return s.UserRepo.Delete(id)
}

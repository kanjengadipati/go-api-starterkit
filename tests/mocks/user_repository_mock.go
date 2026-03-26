package mocks

import (
	"errors"

	"go-auth-app/models"
)

type MockUserRepo struct {
	User *models.User

	FindByEmailErr error
	FindByIDErr    error
}

func (m *MockUserRepo) Create(user *models.User) error {
	m.User = user
	return nil
}

func (m *MockUserRepo) FindByEmail(email string) (*models.User, error) {
	if m.FindByEmailErr != nil {
		return nil, m.FindByEmailErr
	}
	if m.User == nil {
		return nil, errors.New("user not found")
	}
	return m.User, nil
}

func (m *MockUserRepo) FindByID(id uint) (*models.User, error) {
	if m.FindByIDErr != nil {
		return nil, m.FindByIDErr
	}
	if m.User == nil {
		return nil, errors.New("user not found")
	}
	return m.User, nil
}

func (m *MockUserRepo) Update(user *models.User) error {
	m.User = user
	return nil
}

func (m *MockUserRepo) FindAll() ([]models.User, error) {
	if m.User == nil {
		return []models.User{}, nil
	}
	return []models.User{*m.User}, nil
}

func (m *MockUserRepo) FindAllWithFilter(page, limit int, search, role string) ([]models.User, int64, error) {
	// Minimal stub: for unit tests we don't need paging/filter logic.
	if m.User == nil {
		return []models.User{}, 0, nil
	}
	return []models.User{*m.User}, 1, nil
}

func (m *MockUserRepo) Delete(id uint) error {
	// Minimal stub for unit tests.
	// In tests that need this behavior, you can extend it to track calls/ids.
	if m.User != nil && m.User.ID == id {
		m.User = nil
	}
	return nil
}

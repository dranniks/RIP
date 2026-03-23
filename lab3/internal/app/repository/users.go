package repository

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"xrfApp/internal/app/model"
)

type RegisterUserInput struct {
	Login    string
	FullName string
	Password string
}

type AuthInput struct {
	Login    string
	Password string
}

type AuthResult struct {
	UserID        uint
	Login         string
	Role          string
	Authenticated bool
	AuthMode      string
}

func (r *Repository) RegisterUser(input RegisterUserInput) (*model.User, error) {
	login := strings.TrimSpace(input.Login)
	fullName := strings.TrimSpace(input.FullName)
	password := strings.TrimSpace(input.Password)

	if login == "" {
		return nil, fmt.Errorf("%w: login is required", ErrValidation)
	}
	if fullName == "" {
		return nil, fmt.Errorf("%w: full_name is required", ErrValidation)
	}
	if len(password) < 4 {
		return nil, fmt.Errorf("%w: password must have at least 4 chars", ErrValidation)
	}

	var exists int64
	if err := r.db.Model(&model.User{}).Where("login = ?", login).Count(&exists).Error; err != nil {
		return nil, fmt.Errorf("check login uniqueness: %w", err)
	}
	if exists > 0 {
		return nil, fmt.Errorf("%w: login already exists", ErrValidation)
	}

	user := model.User{
		Login:        login,
		FullName:     fullName,
		PasswordHash: hashPassword(password),
		Role:         "creator",
	}
	if err := r.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func (r *Repository) AuthenticateStub(input AuthInput) (*AuthResult, error) {
	login := strings.TrimSpace(input.Login)
	password := strings.TrimSpace(input.Password)
	if login == "" || password == "" {
		return nil, fmt.Errorf("%w: login and password are required", ErrValidation)
	}

	user := model.User{}
	if err := r.db.Where("login = ?", login).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: invalid credentials", ErrValidation)
		}
		return nil, fmt.Errorf("load user by login: %w", err)
	}

	if user.PasswordHash != hashPassword(password) {
		return nil, fmt.Errorf("%w: invalid credentials", ErrValidation)
	}

	return &AuthResult{
		UserID:        user.ID,
		Login:         user.Login,
		Role:          user.Role,
		Authenticated: true,
		AuthMode:      "stub-singleton-no-token",
	}, nil
}

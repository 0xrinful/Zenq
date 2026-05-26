package service

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/0xrinful/Zenq/internal/models"
)

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	return string(hash), err
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *Service) SignUp(ctx context.Context, email, password string) (*models.User, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("service: hash password: %w", err)
	}
	user, err := s.db.CreateUser(email, hash)
	if err != nil {
		return nil, fmt.Errorf("service: create user: %w", err)
	}
	return user, nil
}

func (s *Service) SignIn(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.db.UserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("service: get user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	if !checkPassword(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

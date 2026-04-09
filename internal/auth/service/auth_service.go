package service

import (
	"airline-tracker/internal/user/domain"
	"airline-tracker/internal/user/domain/repository"
	"context"
	"fmt"
	"log/slog"

	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repository repository.UserRepository
}

func NewAuthService(i do.Injector) (*AuthService, error) {
	return &AuthService{
		repository: do.MustInvokeAs[repository.UserRepository](i),
	}, nil
}

func (s *AuthService) GetUser(email, phone, password string) (*domain.User, error) {
	op := "auth_service.Get"
	var u *domain.User
	var err error
	if email != "" {
		u, err = s.repository.GetByEmail(context.Background(), email)
	} else if phone != "" {
		u, err = s.repository.GetByPhone(context.Background(), phone)
	} else {
		err = fmt.Errorf("Empty email and phone")
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		return nil, fmt.Errorf("%s: %s", op, "wrong password")
	}
	return u, nil
}

func (s *AuthService) CreateUser(u *domain.User) error {
	slog.Debug("Pswd before", "pswd", u.Password)
	if encryptPassword(u) != nil { // TODO: move inside NewUser
		return fmt.Errorf("Can't encrypt password")
	}
	slog.Debug("Pswd after", "pswd", u.Password)
	if err := s.repository.Save(context.Background(), u); err != nil {
		return err
	}
	return nil
}

func encryptPassword(u *domain.User) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func (s *AuthService) Exist(email, phone, password string) bool {
	u, err := s.GetUser(email, phone, password)
	if err != nil || u == nil {
		return false
	}
	return true
}

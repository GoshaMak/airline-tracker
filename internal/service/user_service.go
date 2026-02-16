package service

import (
	"airline-ticketing-svc/internal/domain"
	repositories "airline-ticketing-svc/internal/domain/repository"
	"context"
	"fmt"
	"log/slog"

	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repository repositories.UserRepository
}

func NewUserService(i do.Injector) (*UserService, error) {
	return &UserService{
		repository: do.MustInvokeAs[repositories.UserRepository](i),
	}, nil
}

func (s *UserService) GetUser(email string, phone string, password string) (*domain.User, error) {
	op := "user_service.GetUser"
	var u *domain.User
	var err error
	if email != "" {
		u, err = s.repository.GetByEmail(context.Background(), email)
	} else if phone != "" {
		u, err = s.repository.GetByPhone(context.Background(), phone)
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		return nil, fmt.Errorf("%s: %s", op, "wrong password")
	}
	return u, nil
}

func (s *UserService) CreateUser(u *domain.User) error {
	slog.Debug("Pswd before", "pswd", u.Password)
	if encryptPassword(u) != nil {
		return fmt.Errorf("Can't encrypt password")
	}
	slog.Debug("Pswd after", "pswd", u.Password)
	err := s.repository.Save(context.Background(), u)
	return err
}

func encryptPassword(u *domain.User) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

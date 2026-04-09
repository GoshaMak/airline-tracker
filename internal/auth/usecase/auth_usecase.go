package usecase

import (
	"airline-tracker/internal/user/domain"
	"airline-tracker/internal/user/domain/repository"
	"context"
	"fmt"
	"log/slog"

	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	repo repository.UserRepository
}

func NewAuthUsecase(i do.Injector) (*AuthUsecase, error) {
	return &AuthUsecase{
		repo: do.MustInvokeAs[repository.UserRepository](i),
	}, nil
}

func (uc *AuthUsecase) GetUser(email, phone, password string) (*domain.User, error) {
	op := "auth_uc.Get"
	var u *domain.User
	var err error
	if email != "" {
		u, err = uc.repo.GetByEmail(context.Background(), email)
	} else if phone != "" {
		u, err = uc.repo.GetByPhone(context.Background(), phone)
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

func (uc *AuthUsecase) CreateUser(u *domain.User) error {
	slog.Debug("Pswd before", "pswd", u.Password)
	if encryptPassword(u) != nil { // TODO: move inside NewUser
		return fmt.Errorf("Can't encrypt password")
	}
	slog.Debug("Pswd after", "pswd", u.Password)
	if err := uc.repo.Save(context.Background(), u); err != nil {
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

func (uc *AuthUsecase) Exist(email, phone, password string) bool {
	u, err := uc.GetUser(email, phone, password)
	if err != nil || u == nil {
		return false
	}
	return true
}

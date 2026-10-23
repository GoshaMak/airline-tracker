package usecase

import (
	"api/internal/user/domain"
	"api/internal/user/domain/repository"
	"context"
	"errors"
	"fmt"

	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	repo repository.UserRepository
}

func NewAuthUsecase(i do.Injector) (*AuthUsecase, error) {
	return &AuthUsecase{
		repo: do.MustInvoke[repository.UserRepository](i),
	}, nil
}

func (uc *AuthUsecase) GetUser(ctx context.Context, email, password string) (domain.User, error) {
	const op = "AuthUsecase.GetUser"
	u, err := uc.repo.GetUser(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return domain.User{}, ErrUserNotFound
		}
		if errors.Is(err, repository.ErrOperationNotPermitted) {
			return domain.User{}, ErrOperationNotPermitted
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, ErrInvalidPassword
	}

	return u, nil
}

func (uc *AuthUsecase) CreateUser(ctx context.Context, u domain.User) error {
	const op = "AuthUsecase.CreateUser"
	if err := uc.repo.SaveUser(ctx, u); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return ErrUserAlreadyExists
		}
		if errors.Is(err, repository.ErrOperationNotPermitted) {
			return ErrOperationNotPermitted
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// TODO: change to (exists bool, err error)
func (uc *AuthUsecase) Exists(ctx context.Context, email, password string) bool {
	_, err := uc.GetUser(ctx, email, password)
	if err != nil { // WARN:
		return false
	}
	return true
}

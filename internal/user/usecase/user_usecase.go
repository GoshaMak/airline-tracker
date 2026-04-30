package usecase

import (
	flightDomain "airline-tracker/internal/flight/domain"
	"airline-tracker/internal/user/domain"
	"airline-tracker/internal/user/domain/repository"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(i do.Injector) (*UserUsecase, error) {
	return &UserUsecase{
		repo: do.MustInvokeAs[repository.UserRepository](i),
	}, nil
}

func (uc *UserUsecase) GetUser(email, password string) (*domain.User, error) {
	op := "UserUsecase.GetUser"
	var u domain.User
	var err error
	if email != "" {
		u, err = uc.repo.GetUser(context.Background(), email)
	} else {
		err = fmt.Errorf("Empty email and phone")
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, fmt.Errorf("%s: %s", op, "wrong password")
	}
	return &u, nil
}

func (uc *UserUsecase) Exist(email, password string) bool {
	u, err := uc.GetUser(email, password)
	if err != nil || u == nil {
		return false
	}
	return true
}

func (uc *UserUsecase) Subscribe(uid, fid uuid.UUID) error {
	op := "UserUsecase.Subscribe"
	if err := uc.repo.Subscribe(context.Background(), uid, fid); err != nil {
		switch err {
		case repository.ErrUserNotFound:
			return ErrUserNotFound
		case repository.ErrFlightNotFound:
			return ErrFlightNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (uc *UserUsecase) ListFlights(uid uuid.UUID) ([]flightDomain.Flight, error) {
	op := "UserUsecase.ListFlights"
	flights, err := uc.repo.ListFlights(context.Background(), uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return flights, nil
}

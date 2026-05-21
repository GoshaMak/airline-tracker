package usecase

import (
	airportRepository "api/internal/airport/domain/repository"
	flightDomain "api/internal/flight/domain"
	flightRepository "api/internal/flight/domain/repository"
	publisherDomain "api/internal/publisher/domain"
	outboxRepository "api/internal/publisher/domain/repository"
	"api/internal/user/domain"
	userRepository "api/internal/user/domain/repository"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	userRepo   userRepository.UserRepository
	outboxRepo outboxRepository.OutboxRepository
	flightRepo flightRepository.FlightRepository
	gateRepo   airportRepository.GateRepository
}

func NewUserUsecase(i do.Injector) (*UserUsecase, error) {
	return &UserUsecase{
		userRepo:   do.MustInvokeAs[userRepository.UserRepository](i),
		outboxRepo: do.MustInvokeAs[outboxRepository.OutboxRepository](i),
		flightRepo: do.MustInvokeAs[flightRepository.FlightRepository](i),
		gateRepo:   do.MustInvokeAs[airportRepository.GateRepository](i),
	}, nil
}

func (uc *UserUsecase) GetUser(email, password string) (*domain.User, error) {
	const op = "UserUsecase.GetUser"
	var u domain.User
	var err error
	if email != "" {
		u, err = uc.userRepo.GetUser(context.Background(), email)
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

func (uc *UserUsecase) GetUserById(uid uuid.UUID) (domain.User, error) {
	const op = "UserUsecase.GetUserById"
	u, err := uc.userRepo.Exist(context.Background(), uid)
	if err != nil {
		if errors.Is(err, userRepository.ErrUserNotFound) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return u, nil
}

func (uc *UserUsecase) Exist(email, password string) bool {
	u, err := uc.GetUser(email, password)
	if err != nil || u == nil {
		return false
	}
	return true
}

func (uc *UserUsecase) Subscribe(uid, fid uuid.UUID) error {
	const op = "UserUsecase.Subscribe"
	if err := uc.userRepo.Subscribe(context.Background(), uid, fid); err != nil {
		switch err {
		case userRepository.ErrUserNotFound:
			return ErrUserNotFound
		case userRepository.ErrFlightNotFound:
			return ErrFlightNotFound
		case userRepository.ErrUserAlreadySubscribed:
			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	topic := os.Getenv("SUBSCRIPTION_CREATED_TOPIC") // TODO: get it from config?
	payload, err := uc.formPayload(uid, fid)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	ob, err := publisherDomain.NewOutbox(topic, payload)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := uc.outboxRepo.Save(context.Background(), ob); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (uc *UserUsecase) formPayload(uid, fid uuid.UUID) (publisherDomain.Payload, error) {
	const op = "UserUsecase.formPayload"
	u, err := uc.userRepo.Exist(context.Background(), uid)
	if err != nil {
		if errors.Is(err, userRepository.ErrUserNotFound) {
			return publisherDomain.Payload{}, ErrUserNotFound
		}
		return publisherDomain.Payload{}, fmt.Errorf("%s: %w", op, err)
	}

	f, err := uc.flightRepo.Exist(context.Background(), fid)
	if err != nil {
		if errors.Is(err, flightRepository.ErrFlightNotFound) {
			return publisherDomain.Payload{}, ErrFlightNotFound
		}
		return publisherDomain.Payload{}, fmt.Errorf("%s: %w", op, err)
	}

	route, err := uc.flightRepo.GetFlightRoute(context.Background(), fid)
	if err != nil {
		if errors.Is(err, flightRepository.ErrFlightRouteNotFound) {
			return publisherDomain.Payload{}, ErrFlightRouteNotFound
		}
		return publisherDomain.Payload{}, fmt.Errorf("%s: %w", op, err)
	}

	depAirport, err := uc.gateRepo.GetAirportByGateId(context.Background(), route.DepartureGateId)
	if err != nil {
		return publisherDomain.Payload{}, fmt.Errorf("%s: %w", op, err)
	}

	arrAirport, err := uc.gateRepo.GetAirportByGateId(context.Background(), route.ArrivalGateId)
	if err != nil {
		return publisherDomain.Payload{}, fmt.Errorf("%s: %w", op, err)
	}

	p, err := publisherDomain.NewPayload(u.Email, f, depAirport, arrAirport)
	if err != nil {
		return publisherDomain.Payload{}, fmt.Errorf("%s: %w", op, err)
	}
	return p, nil
}

func (uc *UserUsecase) ListFlights(uid uuid.UUID) ([]flightDomain.Flight, error) {
	const op = "UserUsecase.ListFlights"
	flights, err := uc.userRepo.ListFlights(context.Background(), uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return flights, nil
}

package usecase

import (
	gateRepository "api/internal/airport/domain/repository"
	flightRepository "api/internal/flight/domain/repository"
	"api/internal/infra/kafka"
	"api/internal/notification/domain"
	userRepository "api/internal/user/domain/repository"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type NotificationUsecase struct {
	ns         *kafka.NotifySender
	userRepo   userRepository.UserRepository
	flightRepo flightRepository.FlightRepository
	gateRepo   gateRepository.GateRepository
}

func NewNotificationUsecase(i do.Injector) (*NotificationUsecase, error) {
	n := do.MustInvoke[*kafka.NotifySender](i)
	return &NotificationUsecase{
		ns:         n,
		userRepo:   do.MustInvoke[userRepository.UserRepository](i),
		flightRepo: do.MustInvoke[flightRepository.FlightRepository](i),
		gateRepo:   do.MustInvoke[gateRepository.GateRepository](i),
	}, nil
}

func (nu *NotificationUsecase) SendMessage(uid, fid uuid.UUID) error {
	op := "NotificationUsecase.SendMessage"
	u, err := nu.userRepo.Exist(context.Background(), uid)
	if err != nil {
		if errors.Is(err, userRepository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	f, err := nu.flightRepo.Exist(context.Background(), fid)
	if err != nil {
		if errors.Is(err, flightRepository.ErrFlightNotFound) {
			return ErrFlightNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	route, err := nu.flightRepo.GetFlightRoute(context.Background(), fid)
	if err != nil {
		if errors.Is(err, flightRepository.ErrFlightRouteNotFound) {
			return ErrFlightRouteNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	depAirport, err := nu.gateRepo.GetAirportByGateId(context.Background(), route.DepartureGateId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	arrAirport, err := nu.gateRepo.GetAirportByGateId(context.Background(), route.ArrivalGateId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	n, err := domain.NewNotification(u.Email, f, depAirport, arrAirport)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	msg := formMessage(n)
	if err := nu.ns.SendMessage(msg); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func formMessage(n domain.Notification) string {
	return n.String()
}

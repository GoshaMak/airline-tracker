package usecase

import (
	gateRepository "api/internal/airport/domain/repository"
	flightRepository "api/internal/flight/usecase/repository"
	"api/internal/infra/kafka"
	"api/internal/notification/domain"
	"api/internal/notification/usecase/model"
	userRepository "api/internal/user/domain/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"shared/observability"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type NotificationUsecase struct {
	ms       messageSender
	userRepo userRepository.UserRepository
	flightDB flightRepository.FlightDB
	gateRepo gateRepository.GateRepository
}

type messageSender interface {
	SendContext(
		ctx context.Context,
		topic,
		key,
		eventID string,
		version int,
		carrier map[string]string,
		msg []byte,
	) error
}

func NewNotificationUsecase(i do.Injector) (*NotificationUsecase, error) {
	n := do.MustInvoke[*kafka.NotifySender](i)
	return &NotificationUsecase{
		ms:       n,
		userRepo: do.MustInvoke[userRepository.UserRepository](i),
		flightDB: do.MustInvoke[flightRepository.FlightDB](i),
		gateRepo: do.MustInvoke[gateRepository.GateRepository](i),
	}, nil
}

func (nu *NotificationUsecase) SendMessage(ctx context.Context, uid, fid uuid.UUID) error {
	const op = "NotificationUsecase.SendMessage"
	u, err := nu.userRepo.Exist(ctx, uid)
	if err != nil {
		if errors.Is(err, userRepository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	f, err := nu.flightDB.Exist(ctx, fid)
	if err != nil {
		if errors.Is(err, flightRepository.ErrFlightNotFound) {
			return ErrFlightNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	route, err := nu.flightDB.GetFlightRoute(ctx, fid)
	if err != nil {
		if errors.Is(err, flightRepository.ErrFlightRouteNotFound) {
			return ErrFlightRouteNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	depAirport, err := nu.gateRepo.GetAirportByGateId(ctx, route.DepartureGateId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	arrAirport, err := nu.gateRepo.GetAirportByGateId(ctx, route.ArrivalGateId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	n, err := domain.NewNotification(u.Email, f, depAirport, arrAirport)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	nm, err := model.NotificationDomainToModel(n)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	msg, err := json.Marshal(nm)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := nu.ms.SendContext(ctx, "flights", fid.String(), uuid.NewString(), 1, observability.Inject(ctx), msg); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

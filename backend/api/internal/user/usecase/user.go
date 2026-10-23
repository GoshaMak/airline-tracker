package usecase

import (
	airportRepository "api/internal/airport/domain/repository"
	flightDomain "api/internal/flight/domain"
	flightRepository "api/internal/flight/usecase/repository"
	"api/internal/pagination"
	publisherDomain "api/internal/publisher/domain"
	outboxRepository "api/internal/publisher/infra/postgres"
	"api/internal/user/domain"
	userRepository "api/internal/user/domain/repository"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	userRepo   userRepository.UserRepository
	subRepo    userRepository.SubscriptionRepository
	outboxRepo outboxRepository.OutboxRepository
	flightRepo flightRepository.FlightDB
	gateRepo   airportRepository.GateRepository
}

type paginatedSubscriptionRepository interface {
	ListFlightsPage(context.Context, uuid.UUID, pagination.Params) ([]flightDomain.Flight, error)
}

func NewUserUsecase(i do.Injector) (*UserUsecase, error) {
	userRepo := do.MustInvoke[userRepository.UserRepository](i)
	subRepo, ok := userRepo.(userRepository.SubscriptionRepository)
	if !ok {
		subRepo = do.MustInvoke[userRepository.SubscriptionRepository](i)
	}
	return &UserUsecase{
		userRepo:   userRepo,
		subRepo:    subRepo,
		outboxRepo: do.MustInvoke[outboxRepository.OutboxRepository](i),
		flightRepo: do.MustInvoke[flightRepository.FlightDB](i),
		gateRepo:   do.MustInvoke[airportRepository.GateRepository](i),
	}, nil
}

func (uc *UserUsecase) CreateSubscription(ctx context.Context, uid, fid uuid.UUID, notifyBeforeMinutes int64) error {
	const op = "UserUsecase.CreateSubscription"
	var added bool
	err := uc.withTransaction(ctx, func(txctx context.Context) error {
		var err error
		added, err = uc.subRepo.Create(txctx, uid, fid, notifyBeforeMinutes)
		if err != nil || !added {
			return err
		}
		return uc.publishSubscriptionNotification(txctx, uid, fid, notifyBeforeMinutes)
	})
	if err != nil {
		if errors.Is(err, userRepository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		if errors.Is(err, userRepository.ErrFlightNotFound) {
			return ErrFlightNotFound
		}
		if errors.Is(err, userRepository.ErrOperationNotPermitted) {
			return ErrOperationNotPermitted
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (uc *UserUsecase) AddSubscriptionTimers(ctx context.Context, uid, fid uuid.UUID, notifyBeforeMinutes []int64) error {
	const op = "UserUsecase.AddSubscriptionTimers"
	err := uc.withTransaction(ctx, func(txctx context.Context) error {
		added, err := uc.subRepo.AddTimers(txctx, uid, fid, notifyBeforeMinutes)
		if err != nil {
			return err
		}
		for _, minutes := range added {
			if err := uc.publishSubscriptionNotification(txctx, uid, fid, minutes); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, userRepository.ErrSubscriptionNotFound) {
			return ErrSubscriptionNotFound
		}
		if errors.Is(err, userRepository.ErrOperationNotPermitted) {
			return ErrOperationNotPermitted
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (uc *UserUsecase) withTransaction(ctx context.Context, fn func(context.Context) error) error {
	if tx, ok := uc.userRepo.(interface {
		WithTransaction(context.Context, func(context.Context) error) error
	}); ok {
		return tx.WithTransaction(ctx, fn)
	}
	return fn(ctx)
}

func (uc *UserUsecase) DeleteSubscription(ctx context.Context, uid, fid uuid.UUID) error {
	const op = "UserUsecase.DeleteSubscription"
	if err := uc.subRepo.Delete(ctx, uid, fid); err != nil {
		if errors.Is(err, userRepository.ErrOperationNotPermitted) {
			return ErrOperationNotPermitted
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (uc *UserUsecase) publishSubscriptionNotification(ctx context.Context, uid, fid uuid.UUID, notifyBeforeMinutes int64) error {
	const op = "UserUsecase.publishSubscriptionNotification"
	payload, err := uc.formPayload(ctx, uid, fid, notifyBeforeMinutes)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	topic := os.Getenv("SUBSCRIPTION_CREATED_TOPIC")
	ob, err := publisherDomain.NewOutbox(topic, &payload)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := uc.outboxRepo.Save(ctx, ob); err != nil {
		if errors.Is(err, outboxRepository.ErrOperationNotPermitted) {
			return ErrOperationNotPermitted
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	slog.DebugContext(ctx, op+": saved to outbox", "event_id", ob.Id)
	return nil
}

func (uc *UserUsecase) GetUser(ctx context.Context, email, password string) (*domain.User, error) {
	const op = "UserUsecase.GetUser"
	var u domain.User
	var err error
	if email != "" {
		u, err = uc.userRepo.GetUser(ctx, email)
	} else {
		err = fmt.Errorf("Empty email and phone")
	}
	if err != nil {
		if errors.Is(err, userRepository.ErrOperationNotPermitted) {
			return nil, ErrOperationNotPermitted
		}
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, fmt.Errorf("%s: %s", op, "wrong password")
	}
	return &u, nil
}

func (uc *UserUsecase) GetUserById(ctx context.Context, uid uuid.UUID) (domain.User, error) {
	const op = "UserUsecase.GetUserById"
	u, err := uc.userRepo.Exist(ctx, uid)
	if err != nil {
		if errors.Is(err, userRepository.ErrUserNotFound) {
			return domain.User{}, ErrUserNotFound
		}
		if errors.Is(err, userRepository.ErrOperationNotPermitted) {
			return domain.User{}, ErrOperationNotPermitted
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return u, nil
}

func (uc *UserUsecase) Exist(ctx context.Context, email, password string) bool {
	u, err := uc.GetUser(ctx, email, password)
	if err != nil || u == nil {
		return false
	}
	return true
}

func (uc *UserUsecase) Subscribe(ctx context.Context, uid, fid uuid.UUID) error {
	const op = "UserUsecase.Subscribe"
	err := uc.withTransaction(ctx, func(txctx context.Context) error {
		if err := uc.userRepo.Subscribe(txctx, uid, fid); err != nil {
			return err
		}
		topic := os.Getenv("SUBSCRIPTION_CREATED_TOPIC")
		payload, err := uc.formPayload(txctx, uid, fid, 0)
		if err != nil {
			return err
		}
		ob, err := publisherDomain.NewOutbox(topic, &payload)
		if err != nil {
			return err
		}
		return uc.outboxRepo.Save(txctx, ob)
	})
	if err != nil {
		if errors.Is(err, userRepository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		if errors.Is(err, userRepository.ErrFlightNotFound) {
			return ErrFlightNotFound
		}
		if errors.Is(err, userRepository.ErrUserAlreadySubscribed) {
			return nil
		}
		if errors.Is(err, userRepository.ErrOperationNotPermitted) {
			return ErrOperationNotPermitted
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (uc *UserUsecase) formPayload(ctx context.Context, uid, fid uuid.UUID, notifyBeforeMinutes int64) (domain.SubscriptionCreatedPayload, error) {
	const op = "UserUsecase.formPayload"
	u, err := uc.userRepo.Exist(ctx, uid)
	if err != nil {
		if errors.Is(err, userRepository.ErrUserNotFound) {
			return domain.SubscriptionCreatedPayload{}, ErrUserNotFound
		}
		if errors.Is(err, userRepository.ErrOperationNotPermitted) {
			return domain.SubscriptionCreatedPayload{}, ErrOperationNotPermitted
		}
		return domain.SubscriptionCreatedPayload{}, fmt.Errorf("%s: %w", op, err)
	}

	f, err := uc.flightRepo.Exist(ctx, fid)
	if err != nil {
		if errors.Is(err, flightRepository.ErrFlightNotFound) {
			return domain.SubscriptionCreatedPayload{}, ErrFlightNotFound
		}
		if errors.Is(err, flightRepository.ErrOperationNotPermitted) {
			return domain.SubscriptionCreatedPayload{}, ErrOperationNotPermitted
		}
		return domain.SubscriptionCreatedPayload{}, fmt.Errorf("%s: %w", op, err)
	}

	route, err := uc.flightRepo.GetFlightRoute(ctx, fid)
	if err != nil {
		if errors.Is(err, flightRepository.ErrFlightRouteNotFound) {
			return domain.SubscriptionCreatedPayload{}, ErrFlightRouteNotFound
		}
		if errors.Is(err, flightRepository.ErrOperationNotPermitted) {
			return domain.SubscriptionCreatedPayload{}, ErrOperationNotPermitted
		}
		return domain.SubscriptionCreatedPayload{}, fmt.Errorf("%s: %w", op, err)
	}

	depAirport, err := uc.gateRepo.GetAirportByGateId(ctx, route.DepartureGateId)
	if err != nil {
		if errors.Is(err, airportRepository.ErrOperationNotPermitted) {
			return domain.SubscriptionCreatedPayload{}, ErrOperationNotPermitted
		}
		return domain.SubscriptionCreatedPayload{}, fmt.Errorf("%s: %w", op, err)
	}

	arrAirport, err := uc.gateRepo.GetAirportByGateId(ctx, route.ArrivalGateId)
	if err != nil {
		if errors.Is(err, airportRepository.ErrOperationNotPermitted) {
			return domain.SubscriptionCreatedPayload{}, ErrOperationNotPermitted
		}
		return domain.SubscriptionCreatedPayload{}, fmt.Errorf("%s: %w", op, err)
	}

	p, err := domain.NewSubscriptionCreatedPayload(u.Email, f, depAirport, arrAirport, notifyBeforeMinutes)
	if err != nil {
		return domain.SubscriptionCreatedPayload{}, fmt.Errorf("%s: %w", op, err)
	}
	return p, nil
}

func (uc *UserUsecase) ListFlights(ctx context.Context, uid uuid.UUID) ([]flightDomain.Flight, error) {
	const op = "UserUsecase.ListFlights"
	flights, err := uc.subRepo.ListFlights(ctx, uid)
	if err != nil {
		if errors.Is(err, userRepository.ErrOperationNotPermitted) {
			return nil, ErrOperationNotPermitted
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return flights, nil
}

func (uc *UserUsecase) ListFlightsPage(
	ctx context.Context,
	uid uuid.UUID,
	params pagination.Params,
) (pagination.Page[flightDomain.Flight], error) {
	if repo, ok := uc.subRepo.(paginatedSubscriptionRepository); ok {
		flights, err := repo.ListFlightsPage(ctx, uid, params)
		if err != nil {
			if errors.Is(err, userRepository.ErrOperationNotPermitted) {
				return pagination.Page[flightDomain.Flight]{}, ErrOperationNotPermitted
			}
			return pagination.Page[flightDomain.Flight]{}, fmt.Errorf("UserUsecase.ListFlightsPage: %w", err)
		}
		return pagination.FromFetched(flights, params, func(f flightDomain.Flight) uuid.UUID { return f.Id }), nil
	}
	flights, err := uc.ListFlights(ctx, uid)
	if err != nil {
		return pagination.Page[flightDomain.Flight]{}, err
	}
	return pagination.Paginate(flights, params, func(f flightDomain.Flight) uuid.UUID { return f.Id }), nil
}

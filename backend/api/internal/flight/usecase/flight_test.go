package usecase

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/flight/command"
	"api/internal/flight/domain"
	"api/internal/flight/domain/repository"
	publisherDomain "api/internal/publisher/domain"
	publisherRepository "api/internal/publisher/domain/repository"
	userDomain "api/internal/user/domain"
	"context"
	"errors"
	"testing"
	"time"

	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type flightRepoMock struct {
	saveFn            func(ctx context.Context, flight domain.Flight) error
	existFn           func(ctx context.Context, fid uuid.UUID) (domain.Flight, error)
	updateFn          func(ctx context.Context, ufi domain.UpdateFlightInfo) error
	listFlightsFn     func(ctx context.Context) ([]domain.Flight, error)
	getRouteFn        func(ctx context.Context, fid uuid.UUID) (domain.FlightRoute, error)
	listSubscribersFn func(ctx context.Context, fid uuid.UUID) ([]userDomain.User, error)
	getAirportsFn     func(ctx context.Context, fid uuid.UUID) (airportDomain.Airport, airportDomain.Airport, error)
}

func (m *flightRepoMock) Save(ctx context.Context, flight domain.Flight) error {
	return m.saveFn(ctx, flight)
}

func (m *flightRepoMock) Exist(ctx context.Context, fid uuid.UUID) (domain.Flight, error) {
	return m.existFn(ctx, fid)
}

func (m *flightRepoMock) Update(ctx context.Context, ufi domain.UpdateFlightInfo) error {
	return m.updateFn(ctx, ufi)
}

func (m *flightRepoMock) ListFlights(ctx context.Context) ([]domain.Flight, error) {
	return m.listFlightsFn(ctx)
}

func (m *flightRepoMock) GetFlightRoute(ctx context.Context, fid uuid.UUID) (domain.FlightRoute, error) {
	return m.getRouteFn(ctx, fid)
}

func (m *flightRepoMock) ListSubscribers(ctx context.Context, fid uuid.UUID) ([]userDomain.User, error) {
	return m.listSubscribersFn(ctx, fid)
}

func (m *flightRepoMock) GetFlightAirports(
	ctx context.Context,
	fid uuid.UUID,
) (airportDomain.Airport, airportDomain.Airport, error) {
	return m.getAirportsFn(ctx, fid)
}

type outboxRepoMock struct {
	saveFn       func(ctx context.Context, ob publisherDomain.Outbox) error
	listFn       func(ctx context.Context, newPayload func() publisherDomain.Payload) ([]publisherDomain.Outbox, error)
	markAsSentFn func(ctx context.Context, ob publisherDomain.Outbox) error
}

func (m *outboxRepoMock) Save(ctx context.Context, ob publisherDomain.Outbox) error {
	return m.saveFn(ctx, ob)
}

func (m *outboxRepoMock) ListNotSent(
	ctx context.Context,
	newPayload func() publisherDomain.Payload,
) ([]publisherDomain.Outbox, error) {
	return m.listFn(ctx, newPayload)
}

func (m *outboxRepoMock) MarkAsSent(ctx context.Context, ob publisherDomain.Outbox) error {
	return m.markAsSentFn(ctx, ob)
}

func validFlightCommand() command.CreateFlightCommand {
	now := time.Now()
	return command.CreateFlightCommand{
		Flight: command.FlightCommand{
			ScheduledDeparture: now,
			ScheduledArrival:   now.Add(time.Hour),
			Status:             "scheduled",
		},
		AircraftId:         uuid.New(),
		DepartureAiroprtId: uuid.New(),
		ArrivalAiroprtId:   uuid.New(),
		DepartureGateId:    uuid.New(),
		ArrivalGateId:      uuid.New(),
	}
}

func validUpdateFlightCommand(fid uuid.UUID) command.UpdateFlightCommand {
	now := time.Now()
	status := "boarding"
	plan := "AB C-123"
	return command.UpdateFlightCommand{
		FlightId:           fid,
		ScheduledDeparture: &now,
		ScheduledArrival:   new(now.Add(time.Hour)),
		ActualDeparture:    new(now.Add(10 * time.Minute)),
		ActualArrival:      new(now.Add(70 * time.Minute)),
		Status:             &status,
		Plan:               &plan,
	}
}

func TestFlightUsecase_ListFlights(t *testing.T) {
	tests := []struct {
		name    string
		repo    repository.FlightRepository
		wantLen int
		wantErr bool
	}{
		{
			name: "[P] empty list",
			repo: &flightRepoMock{
				listFlightsFn: func(ctx context.Context) ([]domain.Flight, error) {
					return nil, nil
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "[P] non-empty list",
			repo: &flightRepoMock{
				listFlightsFn: func(ctx context.Context) ([]domain.Flight, error) {
					return []domain.Flight{{}, {}}, nil
				},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "[N] repo error",
			repo: &flightRepoMock{
				listFlightsFn: func(ctx context.Context) ([]domain.Flight, error) {
					return nil, errors.New("repo error")
				},
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		allure.Test(t, tt.name, func(allureContext *allure.Context) {
			t := allureContext.T()

			injector := do.New()
			do.Override(injector, func(i do.Injector) (repository.FlightRepository, error) {
				return tt.repo, nil
			})
			do.Override(injector, func(i do.Injector) (publisherRepository.OutboxRepository, error) {
				return &outboxRepoMock{}, nil
			})

			uc, err := NewFlightUsecase(injector)
			if err != nil {
				t.Fatalf("NewFlightUsecase: err = %v", err)
			}

			got, gotErr := uc.ListFlights()
			if gotErr != nil {
				if !tt.wantErr {
					t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListFlights() succeeded unexpectedly")
			}
			if len(got) != tt.wantLen {
				t.Fatalf("got len = %d, want = %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestFlightUsecase_CreateFlight(t *testing.T) {
	tests := []struct {
		name    string
		repo    repository.FlightRepository
		cmd     command.CreateFlightCommand
		wantErr bool
	}{
		{
			name: "[P] valid flight",
			repo: &flightRepoMock{
				saveFn: func(ctx context.Context, flight domain.Flight) error {
					return nil
				},
			},
			cmd:     validFlightCommand(),
			wantErr: false,
		},
		{
			name: "[N] invalid command",
			repo: &flightRepoMock{
				saveFn: func(ctx context.Context, flight domain.Flight) error {
					return nil
				},
			},
			cmd: func() command.CreateFlightCommand {
				cmd := validFlightCommand()
				cmd.Flight.ScheduledArrival = cmd.Flight.ScheduledDeparture.Add(-time.Minute)
				return cmd
			}(),
			wantErr: true,
		},
		{
			name: "[N] repo error",
			repo: &flightRepoMock{
				saveFn: func(ctx context.Context, flight domain.Flight) error {
					return errors.New("repo error")
				},
			},
			cmd:     validFlightCommand(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		allure.Test(t, tt.name, func(allureContext *allure.Context) {
			t := allureContext.T()

			injector := do.New()
			do.Override(injector, func(i do.Injector) (repository.FlightRepository, error) {
				return tt.repo, nil
			})
			do.Override(injector, func(i do.Injector) (publisherRepository.OutboxRepository, error) {
				return &outboxRepoMock{}, nil
			})

			uc, err := NewFlightUsecase(injector)
			if err != nil {
				t.Fatalf("NewFlightUsecase: err = %v", err)
			}

			gotErr := uc.CreateFlight(tt.cmd)
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestFlightUsecase_FlightById(t *testing.T) {
	fid := uuid.New()
	now := time.Now()
	validFlight, err := domain.NewFlight(
		uuid.New(),
		now,
		now.Add(1*time.Hour),
		nil,
		nil,
		"scheduled",
		nil,
		uuid.New(),
		uuid.New(),
		uuid.New(),
		uuid.New(),
	)
	if err != nil {
		t.Fatalf("domain.NewFlight: err = %v", err)
	}

	tests := []struct {
		name    string
		repo    repository.FlightRepository
		wantErr bool
	}{
		{
			name: "[P] found",
			repo: &flightRepoMock{
				existFn: func(ctx context.Context, fid uuid.UUID) (domain.Flight, error) {
					return validFlight, nil
				},
			},
			wantErr: false,
		},
		{
			name: "[N] not found",
			repo: &flightRepoMock{
				existFn: func(ctx context.Context, fid uuid.UUID) (domain.Flight, error) {
					return domain.Flight{}, repository.ErrFlightNotFound
				},
			},
			wantErr: true,
		},
		{
			name: "[N] repo error",
			repo: &flightRepoMock{
				existFn: func(ctx context.Context, fid uuid.UUID) (domain.Flight, error) {
					return domain.Flight{}, errors.New("repo error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		allure.Test(t, tt.name, func(allureContext *allure.Context) {
			t := allureContext.T()

			injector := do.New()
			do.Override(injector, func(i do.Injector) (repository.FlightRepository, error) {
				return tt.repo, nil
			})
			do.Override(injector, func(i do.Injector) (publisherRepository.OutboxRepository, error) {
				return &outboxRepoMock{}, nil
			})

			uc, err := NewFlightUsecase(injector)
			if err != nil {
				t.Fatalf("NewFlightUsecase: err = %v", err)
			}

			_, gotErr := uc.FlightById(fid)
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestFlightUsecase_UpdateFlight(t *testing.T) {
	fid := uuid.New()
	tests := []struct {
		name    string
		repo    repository.FlightRepository
		wantErr bool
	}{
		{
			name: "[P] updated",
			repo: &flightRepoMock{
				updateFn: func(ctx context.Context, ufi domain.UpdateFlightInfo) error {
					return nil
				},
				listSubscribersFn: func(ctx context.Context, fid uuid.UUID) ([]userDomain.User, error) {
					return nil, nil
				},
				getAirportsFn: func(ctx context.Context, fid uuid.UUID) (airportDomain.Airport, airportDomain.Airport, error) {
					return airportDomain.Airport{}, airportDomain.Airport{}, nil
				},
			},
			wantErr: false,
		},
		{
			name: "[N] not found",
			repo: &flightRepoMock{
				updateFn: func(ctx context.Context, ufi domain.UpdateFlightInfo) error {
					return repository.ErrFlightNotFound
				},
			},
			wantErr: true,
		},
		{
			name: "[N] repo error",
			repo: &flightRepoMock{
				updateFn: func(ctx context.Context, ufi domain.UpdateFlightInfo) error {
					return errors.New("repo error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		allure.Test(t, tt.name, func(allureContext *allure.Context) {
			t := allureContext.T()

			injector := do.New()
			do.Override(injector, func(i do.Injector) (repository.FlightRepository, error) {
				return tt.repo, nil
			})
			do.Override(injector, func(i do.Injector) (publisherRepository.OutboxRepository, error) {
				return &outboxRepoMock{
					saveFn: func(ctx context.Context, ob publisherDomain.Outbox) error {
						return nil
					},
				}, nil
			})

			uc, err := NewFlightUsecase(injector)
			if err != nil {
				t.Fatalf("NewFlightUsecase: err = %v", err)
			}

			gotErr := uc.UpdateFlight(validUpdateFlightCommand(fid))
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
			}
		})
	}
}

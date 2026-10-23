package usecase

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/flight/command"
	"api/internal/flight/domain"
	flightPostgres "api/internal/flight/infra/postgres"
	flightRedis "api/internal/flight/infra/redis"
	redisModel "api/internal/flight/infra/redis/model"
	"api/internal/flight/usecase/repository"
	flightRepository "api/internal/flight/usecase/repository"
	"api/internal/pagination"
	publisherDomain "api/internal/publisher/domain"
	publisherRepository "api/internal/publisher/infra/postgres"
	userDomain "api/internal/user/domain"
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"shared/common"
	"strings"
	"testing"
	"time"

	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v6"
	"github.com/samber/do/v2"
)

type flightRepoMock struct {
	repository.FlightDB
	saveFn            func(context.Context, domain.Flight) error
	existFn           func(context.Context, uuid.UUID) (domain.Flight, error)
	updateFn          func(context.Context, domain.UpdateFlightInfo) error
	listFlightsFn     func(context.Context) ([]domain.Flight, error)
	listSubscribersFn func(context.Context, uuid.UUID) ([]userDomain.User, error)
	getAirportsFn     func(context.Context, uuid.UUID) (airportDomain.Airport, airportDomain.Airport, error)
}

func (m *flightRepoMock) Save(ctx context.Context, f domain.Flight) error {
	return m.saveFn(ctx, f)
}

func (m *flightRepoMock) Exist(ctx context.Context, id uuid.UUID) (domain.Flight, error) {
	return m.existFn(ctx, id)
}

func (m *flightRepoMock) Update(ctx context.Context, u domain.UpdateFlightInfo) error {
	return m.updateFn(ctx, u)
}

func (m *flightRepoMock) ListFlights(ctx context.Context) ([]domain.Flight, error) {
	return m.listFlightsFn(ctx)
}

func (m *flightRepoMock) ListSubscribers(ctx context.Context, id uuid.UUID) ([]userDomain.User, error) {
	return m.listSubscribersFn(ctx, id)
}

func (m *flightRepoMock) GetFlightAirports(ctx context.Context, id uuid.UUID) (airportDomain.Airport, airportDomain.Airport, error) {
	return m.getAirportsFn(ctx, id)
}

type outboxRepoMock struct {
	publisherRepository.OutboxRepository
	saveFn func(context.Context, publisherDomain.Outbox) error
}

func (m *outboxRepoMock) Save(ctx context.Context, ob publisherDomain.Outbox) error {
	return m.saveFn(ctx, ob)
}

func validFlightCommand() command.CreateFlightCommand {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	return command.CreateFlightCommand{
		Flight: command.FlightCommand{
			ScheduledDeparture: now,
			ScheduledArrival:   now.Add(time.Hour),
			Status:             "scheduled",
		},
		AircraftId: uuid.New(), DepartureAiroprtId: uuid.New(), ArrivalAiroprtId: uuid.New(),
		DepartureGateId: uuid.New(), ArrivalGateId: uuid.New(),
	}
}

func validUpdateFlightCommand(id uuid.UUID) command.UpdateFlightCommand {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	return command.UpdateFlightCommand{
		FlightId:           id,
		ScheduledDeparture: &now,
		ScheduledArrival:   new(now.Add(time.Hour)),
		ActualDeparture:    new(now.Add(10 * time.Minute)),
		ActualArrival:      new(now.Add(70 * time.Minute)),
		Status:             new("boarding"), Plan: new("AB C-123"),
	}
}

func flightFromCommand(cmd command.CreateFlightCommand) domain.Flight {
	var plan *domain.FlightPlan
	if cmd.Flight.Plan != nil {
		plan = new(domain.FlightPlan(*cmd.Flight.Plan))
	}
	return domain.Flight{
		AircraftId:         cmd.AircraftId,
		ScheduledDeparture: cmd.Flight.ScheduledDeparture,
		ScheduledArrival:   cmd.Flight.ScheduledArrival,
		ActualDeparture:    cmd.Flight.ActualDeparture,
		ActualArrival:      cmd.Flight.ActualArrival,
		Status:             domain.Scheduled,
		Plan:               plan,
		DepartureAirportId: cmd.DepartureAiroprtId,
		ArrivalAirportId:   cmd.ArrivalAiroprtId,
		DepartureGateId:    cmd.DepartureGateId,
		ArrivalGateId:      cmd.ArrivalGateId,
	}
}

func newUsecaseFixture(repo repository.FlightDB) *FlightUsecase {
	return &FlightUsecase{repo: repo}
}

func newClassicalUsecaseFixture(t *testing.T) (*FlightUsecase, pgxmock.PgxPoolIface, redismock.ClientMock) {
	t.Helper()
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("create pgxmock pool: %v", err)
	}
	t.Cleanup(pool.Close)
	client, redisMock := redismock.NewClientMock()
	t.Cleanup(func() { _ = client.Close() })

	injector := do.New()
	do.ProvideValue(injector, pool)
	do.ProvideValue(injector, client)
	db, err := flightPostgres.NewPostgresDB(injector)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := flightRedis.NewRedisDB(injector)
	if err != nil {
		t.Fatal(err)
	}
	do.ProvideValue(injector, db)
	do.ProvideValue(injector, cache)
	repo, err := flightRepository.NewFlightRepository(injector)
	if err != nil {
		t.Fatal(err)
	}
	return newUsecaseFixture(repo), pool, redisMock
}

func TestFlightUsecase_CreateFlight_Classical(t *testing.T) {
	allure.Test(t, "positive: real repository saves flight", func(ctx *allure.Context) {
		t := ctx.T()

		uc, pool, cache := newClassicalUsecaseFixture(t)
		defer func() {
			if err := pool.ExpectationsWereMet(); err != nil {
				t.Errorf("database expectations: %v", err)
			}
			if err := cache.ExpectationsWereMet(); err != nil {
				t.Errorf("redis expectations: %v", err)
			}
		}()
		cmd := validFlightCommand()
		cmd.Flight.Plan = new("DCT SVO LED")
		var flightID uuid.UUID
		captureID := pgxmock.ArgumentFunc(func(value any) bool {
			id, ok := value.(uuid.UUID)
			if !ok || id == uuid.Nil {
				return false
			}
			flightID = id
			return true
		})
		pool.ExpectExec(regexp.QuoteMeta("select add_flight($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)")).
			WithArgs(captureID, cmd.Flight.ScheduledDeparture, cmd.Flight.ScheduledArrival, cmd.Flight.Status,
				cmd.Flight.Plan, cmd.AircraftId, cmd.DepartureAiroprtId, cmd.ArrivalAiroprtId,
				cmd.DepartureGateId, cmd.ArrivalGateId).
			WillReturnResult(pgxmock.NewResult("SELECT", 1))

		const generatedKey = "<created flight key>"
		matchFlightKey := func(expected, actual []any) error {
			expected = append([]any(nil), expected...)
			for i, value := range expected {
				if key, ok := value.(string); ok && key == generatedKey {
					expected[i] = "flight:" + flightID.String()
				}
			}
			if !reflect.DeepEqual(expected, actual) {
				return fmt.Errorf("redis args=%v, want=%v", actual, expected)
			}
			return nil
		}
		cache.ExpectTxPipeline()
		cache.CustomMatch(matchFlightKey).ExpectHSet(generatedKey, redisModel.FlightModel{
			AircraftId:         cmd.AircraftId.String(),
			ScheduledDeparture: cmd.Flight.ScheduledDeparture,
			ScheduledArrival:   cmd.Flight.ScheduledArrival,
			Status:             cmd.Flight.Status,
			Plan:               cmd.Flight.Plan,
			DepartureAirportId: cmd.DepartureAiroprtId.String(),
			ArrivalAirportId:   cmd.ArrivalAiroprtId.String(),
			DepartureGateId:    cmd.DepartureGateId.String(),
			ArrivalGateId:      cmd.ArrivalGateId.String(),
		}).SetVal(1)
		cache.CustomMatch(matchFlightKey).ExpectSAdd("flight:ids", generatedKey).SetVal(1)
		cache.ExpectTxPipelineExec()

		err := uc.CreateFlight(t.Context(), cmd)

		if err != nil {
			t.Fatalf("create flight: %v", err)
		}
		if flightID == uuid.Nil {
			t.Fatal("created flight has no ID")
		}
	})
}

func TestFlightUsecase_ListFlights(t *testing.T) {
	failure := errors.New("repository unavailable")
	f := domain.Flight{Id: uuid.New()}
	for _, tt := range []struct {
		name    string
		flights []domain.Flight
		failure error
	}{
		{
			name: "positive: empty list",
		},
		{
			name:    "positive: returns flights",
			flights: []domain.Flight{f},
		},
		{
			name:    "negative: preserves repository error",
			failure: failure,
		},
	} {
		allure.Test(t, tt.name, func(ctx *allure.Context) {
			t := ctx.T()
			calls := 0
			repo := &flightRepoMock{listFlightsFn: func(context.Context) ([]domain.Flight, error) { calls++; return tt.flights, tt.failure }}
			uc := newUsecaseFixture(repo)
			got, err := uc.ListFlights(t.Context())
			if !errors.Is(err, tt.failure) || len(got) != len(tt.flights) || calls != 1 {
				t.Fatalf("got=%+v, err=%v, calls=%d", got, err, calls)
			}
			for i, flight := range got {
				if flight.Id != tt.flights[i].Id {
					t.Fatalf("flight ID=%s, want=%s", flight.Id, tt.flights[i].Id)
				}
			}
		})
	}
}

func TestFlightUsecase_CreateFlight(t *testing.T) {
	failure := errors.New("repository unavailable")
	for _, tt := range []struct {
		name    string
		invalid bool
		failure error
	}{
		{
			name: "positive: saves complete flight",
		},
		{
			name:    "negative: invalid dates prevent save",
			invalid: true,
		},
		{
			name:    "negative: preserves repository error",
			failure: failure,
		},
	} {
		allure.Test(t, tt.name, func(ctx *allure.Context) {
			t := ctx.T()
			cmd := validFlightCommand()
			cmd.Flight.ActualDeparture = new(cmd.Flight.ScheduledDeparture.Add(time.Minute))
			cmd.Flight.ActualArrival = new(cmd.Flight.ScheduledArrival.Add(time.Minute))
			cmd.Flight.Plan = new("DCT SVO LED")
			if tt.invalid {
				cmd.Flight.ScheduledArrival = cmd.Flight.ScheduledDeparture.Add(-time.Minute)
			}
			var saved []domain.Flight
			repo := &flightRepoMock{saveFn: func(_ context.Context, f domain.Flight) error { saved = append(saved, f); return tt.failure }}
			uc := newUsecaseFixture(repo)
			err := uc.CreateFlight(t.Context(), cmd)
			if tt.invalid {
				if err == nil || !strings.Contains(err.Error(), "departure after arrival") || len(saved) != 0 {
					t.Fatalf("err=%v, saved=%v", err, saved)
				}
				return
			}
			if !errors.Is(err, tt.failure) || len(saved) != 1 {
				t.Fatalf("err=%v, saved=%v", err, saved)
			}
			want := flightFromCommand(cmd)
			if saved[0].Id == uuid.Nil {
				t.Fatal("saved flight has no ID")
			}
			want.Id = saved[0].Id
			if !reflect.DeepEqual(saved[0], want) {
				t.Fatalf("saved=%+v, want=%+v", saved[0], want)
			}
		})
	}
}

func TestFlightUsecase_FlightById(t *testing.T) {
	failure := errors.New("repository unavailable")
	for _, tt := range []struct {
		name             string
		repoErr, wantErr error
	}{
		{
			name: "positive: returns flight",
		},
		{
			name:    "negative: maps missing flight",
			repoErr: repository.ErrFlightNotFound,
			wantErr: ErrFlightNotFound,
		},
		{
			name:    "negative: preserves repository error",
			repoErr: failure,
			wantErr: failure,
		},
	} {
		allure.Test(t, tt.name, func(ctx *allure.Context) {
			t := ctx.T()
			f := domain.Flight{Id: uuid.New()}
			calls := 0
			var requested uuid.UUID
			repo := &flightRepoMock{existFn: func(_ context.Context, id uuid.UUID) (domain.Flight, error) {
				calls++
				requested = id
				if tt.repoErr != nil {
					return domain.Flight{}, tt.repoErr
				}
				return f, nil
			}}
			uc := newUsecaseFixture(repo)
			got, err := uc.FlightById(t.Context(), f.Id)
			wantID := f.Id
			if tt.wantErr != nil {
				wantID = uuid.Nil
			}
			if !errors.Is(err, tt.wantErr) || got.Id != wantID || requested != f.Id || calls != 1 {
				t.Fatalf("got=%+v, err=%v, requested=%s, calls=%d", got, err, requested, calls)
			}
		})
	}
}

type updateFlightFixture struct {
	uc          *FlightUsecase
	cmd         command.UpdateFlightCommand
	updated     []domain.UpdateFlightInfo
	queued      chan publisherDomain.Outbox
	wantUpdate  domain.UpdateFlightInfo
	wantPayload *domain.FlightUpdatedPayload
}

func newUpdateFlightFixture(t *testing.T, repoErr error) *updateFlightFixture {
	t.Helper()
	t.Setenv("FLIGHT_UPDATED_TOPIC", "flight.updated")
	cmd := validUpdateFlightCommand(uuid.New())
	email, err := common.NewEmail("traveler@example.com")
	if err != nil {
		t.Fatal(err)
	}
	fixture := &updateFlightFixture{
		cmd:    cmd,
		queued: make(chan publisherDomain.Outbox, 1),
		wantUpdate: domain.UpdateFlightInfo{
			FlightId:           cmd.FlightId,
			ScheduledDeparture: cmd.ScheduledDeparture,
			ScheduledArrival:   cmd.ScheduledArrival,
			ActualDeparture:    cmd.ActualDeparture,
			ActualArrival:      cmd.ActualArrival,
			Status:             cmd.Status,
			Plan:               cmd.Plan,
		},
		wantPayload: &domain.FlightUpdatedPayload{
			FlightId:              cmd.FlightId,
			Users:                 []common.Email{email},
			DepartureAirportTitle: airportDomain.Title("Sheremetyevo"),
			ArrivalAirportTitle:   airportDomain.Title("Pulkovo"),
			ScheduledDeparture:    cmd.ScheduledDeparture,
			ScheduledArrival:      cmd.ScheduledArrival,
			ActualDeparture:       cmd.ActualDeparture,
			ActualArrival:         cmd.ActualArrival,
			Status:                cmd.Status,
			Plan:                  cmd.Plan,
		},
	}
	repo := &flightRepoMock{
		updateFn: func(_ context.Context, u domain.UpdateFlightInfo) error {
			fixture.updated = append(fixture.updated, u)
			return repoErr
		},
		listSubscribersFn: func(_ context.Context, id uuid.UUID) ([]userDomain.User, error) {
			if id != cmd.FlightId {
				return nil, errors.New("incorrect flight ID")
			}
			return []userDomain.User{{Email: email}}, nil
		},
		getAirportsFn: func(_ context.Context, id uuid.UUID) (airportDomain.Airport, airportDomain.Airport, error) {
			if id != cmd.FlightId {
				return airportDomain.Airport{}, airportDomain.Airport{}, errors.New("incorrect flight ID")
			}
			return airportDomain.Airport{Title: airportDomain.Title("Sheremetyevo")},
				airportDomain.Airport{Title: airportDomain.Title("Pulkovo")}, nil
		},
	}
	fixture.uc = newUsecaseFixture(repo)
	fixture.uc.outboxRepo = &outboxRepoMock{
		saveFn: func(_ context.Context, ob publisherDomain.Outbox) error {
			fixture.queued <- ob
			return nil
		},
	}
	return fixture
}

func (f *updateFlightFixture) assertUpdate(t *testing.T) {
	t.Helper()
	if !reflect.DeepEqual(f.updated, []domain.UpdateFlightInfo{f.wantUpdate}) {
		t.Fatalf("updated=%+v, want=%+v", f.updated, f.wantUpdate)
	}
}

func (f *updateFlightFixture) assertNotification(t *testing.T) {
	t.Helper()
	select {
	case ob := <-f.queued:
		if ob.Topic != "flight.updated" || ob.Id == uuid.Nil || ob.SentAt != nil || !reflect.DeepEqual(ob.Payload, f.wantPayload) {
			t.Fatalf("outbox=%+v, payload=%+v", ob, ob.Payload)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("notification was not saved to outbox")
	}
}

func TestFlightUsecase_UpdateFlight(t *testing.T) {
	failure := errors.New("repository unavailable")
	for _, tt := range []struct {
		name             string
		repoErr, wantErr error
	}{
		{
			name: "positive: updates and queues notification",
		},
		{
			name:    "negative: maps missing flight",
			repoErr: repository.ErrFlightNotFound,
			wantErr: ErrFlightNotFound,
		},
		{
			name:    "negative: preserves repository error",
			repoErr: failure,
			wantErr: failure,
		},
	} {
		allure.Test(t, tt.name, func(ctx *allure.Context) {
			t := ctx.T()
			fixture := newUpdateFlightFixture(t, tt.repoErr)

			err := fixture.uc.UpdateFlight(t.Context(), fixture.cmd)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err=%v, want=%v", err, tt.wantErr)
			}
			fixture.assertUpdate(t)
			if tt.wantErr == nil {
				fixture.assertNotification(t)
			}
		})
	}
}

type paginatedRepoMock struct {
	*flightRepoMock
	pageFn func(context.Context, pagination.Params) ([]domain.Flight, error)
}

func (m *paginatedRepoMock) ListFlightsPage(ctx context.Context, p pagination.Params) ([]domain.Flight, error) {
	return m.pageFn(ctx, p)
}

func TestFlightUsecase_ListFlightsPage(t *testing.T) {
	failure := errors.New("repository unavailable")
	for _, tt := range []struct {
		name      string
		paginated bool
		failure   error
	}{
		{
			name:      "positive: database page",
			paginated: true,
		},
		{
			name:      "negative: database page error",
			paginated: true,
			failure:   failure,
		},
		{
			name: "positive: fallback sorts and filters",
		},
		{
			name:    "negative: fallback error",
			failure: failure,
		},
	} {
		allure.Test(t, tt.name, func(ctx *allure.Context) {
			t := ctx.T()
			first := domain.Flight{Id: uuid.MustParse("00000000-0000-0000-0000-000000000001")}
			second := domain.Flight{Id: uuid.MustParse("00000000-0000-0000-0000-000000000002")}
			third := domain.Flight{Id: uuid.MustParse("00000000-0000-0000-0000-000000000003")}
			params := pagination.Params{Limit: 1, Cursor: &first.Id}
			calls := 0
			mock := &flightRepoMock{listFlightsFn: func(context.Context) ([]domain.Flight, error) {
				calls++
				return []domain.Flight{third, first, second}, tt.failure
			}}
			var repo repository.FlightDB = mock
			if tt.paginated {
				repo = &paginatedRepoMock{flightRepoMock: mock, pageFn: func(_ context.Context, p pagination.Params) ([]domain.Flight, error) {
					calls++
					if p != params {
						t.Fatalf("params=%+v, want=%+v", p, params)
					}
					return []domain.Flight{second, third}, tt.failure
				}}
			}
			uc := newUsecaseFixture(repo)
			got, err := uc.ListFlightsPage(t.Context(), params)
			if !errors.Is(err, tt.failure) || calls != 1 {
				t.Fatalf("err=%v, calls=%d", err, calls)
			}
			if tt.failure != nil {
				if got.Items != nil || got.NextCursor != nil || got.HasMore {
					t.Fatalf("page on error=%+v", got)
				}
				return
			}
			if !reflect.DeepEqual(got.Items, []domain.Flight{second}) || !got.HasMore || got.NextCursor == nil {
				t.Fatalf("page=%+v", got)
			}
			next, err := pagination.Parse("1", *got.NextCursor)
			if err != nil || next.Cursor == nil || *next.Cursor != second.Id {
				t.Fatalf("next=%+v, err=%v", next, err)
			}
		})
	}
}

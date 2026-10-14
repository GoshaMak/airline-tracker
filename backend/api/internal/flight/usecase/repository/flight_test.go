package repository

import (
	airportDomain "api/internal/airport/domain"
	"api/internal/flight/domain"
	flightDomainRepository "api/internal/flight/domain/repository"
	"api/internal/flight/infra/postgres"
	"api/internal/flight/infra/redis"
	userDomain "api/internal/user/domain"
	"context"
	"errors"
	"testing"

	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/google/uuid"
)

type dbStub struct {
	flightDomainRepository.FlightRepository
	flight                                        domain.Flight
	flights                                       []domain.Flight
	route                                         domain.FlightRoute
	subs                                          []userDomain.User
	dep, arr                                      airportDomain.Airport
	saved                                         domain.Flight
	updated                                       domain.UpdateFlightInfo
	requested                                     uuid.UUID
	err                                           error
	saveCalls, existCalls, updateCalls, listCalls int
}

func (s *dbStub) Save(_ context.Context, f domain.Flight) error {
	s.saveCalls++
	s.saved = f
	return s.err
}

func (s *dbStub) Exist(_ context.Context, id uuid.UUID) (domain.Flight, error) {
	s.requested = id
	s.existCalls++
	return s.flight, s.err
}

func (s *dbStub) Update(_ context.Context, u domain.UpdateFlightInfo) error {
	s.updated = u
	s.updateCalls++
	return s.err
}

func (s *dbStub) ListFlights(context.Context) ([]domain.Flight, error) {
	s.listCalls++
	return s.flights, s.err
}

func (s *dbStub) GetFlightRoute(_ context.Context, id uuid.UUID) (domain.FlightRoute, error) {
	s.requested = id
	return s.route, s.err
}

func (s *dbStub) ListSubscribers(_ context.Context, id uuid.UUID) ([]userDomain.User, error) {
	s.requested = id
	return s.subs, s.err
}

func (s *dbStub) GetFlightAirports(_ context.Context, id uuid.UUID) (airportDomain.Airport, airportDomain.Airport, error) {
	s.requested = id
	return s.dep, s.arr, s.err
}

type cacheStub struct {
	cache
	flight                                       domain.Flight
	flights                                      []domain.Flight
	err                                          error
	saveCalls, getCalls, flushCalls, updateCalls int
	saved                                        domain.Flight
	updated                                      domain.UpdateFlightInfo
	requested                                    uuid.UUID
}

func (s *cacheStub) SaveFlight(_ context.Context, f domain.Flight) error {
	s.saveCalls++
	s.saved = f
	return s.err
}

func (s *cacheStub) SaveFlights(context.Context, []domain.Flight) error {
	s.saveCalls++
	return s.err
}

func (s *cacheStub) GetFlightById(_ context.Context, id uuid.UUID) (domain.Flight, error) {
	s.requested = id
	s.getCalls++
	return s.flight, s.err
}

func (s *cacheStub) UpdateFlight(_ context.Context, u domain.UpdateFlightInfo) error {
	s.updateCalls++
	s.updated = u
	return s.err
}

func (s *cacheStub) GetFlights(context.Context) ([]domain.Flight, error) {
	s.getCalls++
	return s.flights, s.err
}

func (s *cacheStub) FlushFlights(context.Context) error {
	s.flushCalls++
	return nil
}

func newFixture() (*flightRepository, *dbStub, *cacheStub) {
	db, cache := &dbStub{}, &cacheStub{}
	return &flightRepository{db: db, rd: cache}, db, cache
}

func TestFlightRepository_Save(t *testing.T) {
	allure.Test(t, "positive: writes database and cache", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, cache := newFixture()
		f := domain.Flight{Id: uuid.New()}
		err := repo.Save(context.Background(), f)
		if err != nil || db.saveCalls != 1 || cache.saveCalls != 1 ||
			db.saved != f || cache.saved != f {
			t.Fatalf("err=%v, db=%d, cache=%d", err, db.saveCalls, cache.saveCalls)
		}
	})

	allure.Test(t, "negative: database failure prevents cache write", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, cache := newFixture()
		failure := errors.New("database unavailable")
		db.err = failure
		err := repo.Save(context.Background(), domain.Flight{})
		if !errors.Is(err, failure) || cache.saveCalls != 0 {
			t.Fatalf("err=%v, cache=%d", err, cache.saveCalls)
		}
	})
}

func TestFlightRepository_Exist(t *testing.T) {
	allure.Test(t, "positive: retrieves persisted flight", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, cache := newFixture()
		fid := uuid.New()
		db.flight = domain.Flight{Id: fid}
		got, err := repo.Exist(context.Background(), fid)
		if err != nil || got.Id != fid || db.requested != fid || cache.requested != fid || db.existCalls != 1 || cache.getCalls != 1 {
			t.Fatalf("got=%+v, err=%v, db=%d, cache=%d", got, err, db.existCalls, cache.getCalls)
		}
	})

	allure.Test(t, "negative: missing flight is mapped", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, _ := newFixture()
		db.err = postgres.ErrFlightNotFound
		_, err := repo.Exist(context.Background(), uuid.New())
		if !errors.Is(err, flightDomainRepository.ErrFlightNotFound) {
			t.Fatal(err)
		}
	})
}

func TestFlightRepository_Update(t *testing.T) {
	allure.Test(t, "positive: updates database and cache", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, cache := newFixture()
		status := "boarding"
		update := domain.UpdateFlightInfo{FlightId: uuid.New(), Status: &status}
		err := repo.Update(context.Background(), update)
		if err != nil || db.updateCalls != 1 || cache.updateCalls != 1 ||
			db.updated != update || cache.updated != update {
			t.Fatalf("err=%v, db=%+v, cache=%+v", err, db, cache)
		}
	})

	allure.Test(t, "negative: missing flight", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, cache := newFixture()
		db.err = postgres.ErrFlightNotFound
		err := repo.Update(context.Background(), domain.UpdateFlightInfo{FlightId: uuid.New()})
		if !errors.Is(err, flightDomainRepository.ErrFlightNotFound) || cache.updateCalls != 0 {
			t.Fatalf("err=%v, cache updates=%d", err, cache.updateCalls)
		}
	})
}

func TestFlightRepository_ListFlights(t *testing.T) {
	allure.Test(t, "positive: returns cached flights", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, cache := newFixture()
		fid := uuid.New()
		cache.flights = []domain.Flight{{Id: fid}}
		got, err := repo.ListFlights(context.Background())
		if err != nil || len(got) != 1 || got[0].Id != fid || db.listCalls != 0 {
			t.Fatalf("got=%+v, err=%v, db=%d", got, err, db.listCalls)
		}
	})

	allure.Test(t, "negative: database fallback fails", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, cache := newFixture()
		cache.err = redis.ErrCacheEmpty
		failure := errors.New("database unavailable")
		db.err = failure
		_, err := repo.ListFlights(context.Background())
		if !errors.Is(err, failure) || db.listCalls != 1 {
			t.Fatalf("err=%v, db=%d", err, db.listCalls)
		}
	})
}

func TestFlightRepository_GetFlightRoute(t *testing.T) {
	allure.Test(t, "positive: returns route", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, _ := newFixture()
		rid := uuid.New()
		db.route = domain.FlightRoute{Id: rid}
		fid := uuid.New()
		got, err := repo.GetFlightRoute(context.Background(), fid)
		if err != nil || got.Id != rid || db.requested != fid {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})

	allure.Test(t, "negative: missing route is mapped", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, _ := newFixture()
		db.err = postgres.ErrFlightRouteNotFound
		_, err := repo.GetFlightRoute(context.Background(), uuid.New())
		if !errors.Is(err, flightDomainRepository.ErrFlightRouteNotFound) {
			t.Fatal(err)
		}
	})
}

func TestFlightRepository_ListSubscribers(t *testing.T) {
	allure.Test(t, "positive: returns subscribers", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, _ := newFixture()
		uid := uuid.New()
		db.subs = []userDomain.User{{Id: uid}}
		fid := uuid.New()
		got, err := repo.ListSubscribers(context.Background(), fid)
		if err != nil || len(got) != 1 || got[0].Id != uid || db.requested != fid {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})

	allure.Test(t, "negative: database error", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, _ := newFixture()
		failure := errors.New("database unavailable")
		db.err = failure
		_, err := repo.ListSubscribers(context.Background(), uuid.New())
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

func TestFlightRepository_GetFlightAirports(t *testing.T) {
	allure.Test(t, "positive: returns both airports", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, _ := newFixture()
		db.dep = airportDomain.Airport{ID: uuid.New()}
		db.arr = airportDomain.Airport{ID: uuid.New()}
		fid := uuid.New()
		dep, arr, err := repo.GetFlightAirports(context.Background(), fid)
		if err != nil || dep.ID != db.dep.ID ||
			arr.ID != db.arr.ID || db.requested != fid {
			t.Fatalf("dep=%+v, arr=%+v, err=%v", dep, arr, err)
		}
	})

	allure.Test(t, "negative: database error", func(allureCtx *allure.Context) {
		t := allureCtx.T()

		repo, db, _ := newFixture()
		failure := errors.New("database unavailable")
		db.err = failure
		_, _, err := repo.GetFlightAirports(context.Background(), uuid.New())
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

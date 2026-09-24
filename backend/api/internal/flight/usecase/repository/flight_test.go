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

	"github.com/google/uuid"
)

// dbStub and cacheStub record boundary interactions without external services.
type dbStub struct {
	flightDomainRepository.FlightRepository
	flight                                        domain.Flight
	flights                                       []domain.Flight
	route                                         domain.FlightRoute
	subs                                          []userDomain.User
	dep, arr                                      airportDomain.Airport
	err                                           error
	saveCalls, existCalls, updateCalls, listCalls int
}

func (s *dbStub) Save(context.Context, domain.Flight) error { s.saveCalls++; return s.err }
func (s *dbStub) Exist(context.Context, uuid.UUID) (domain.Flight, error) {
	s.existCalls++
	return s.flight, s.err
}
func (s *dbStub) Update(context.Context, domain.UpdateFlightInfo) error {
	s.updateCalls++
	return s.err
}
func (s *dbStub) ListFlights(context.Context) ([]domain.Flight, error) {
	s.listCalls++
	return s.flights, s.err
}
func (s *dbStub) GetFlightRoute(context.Context, uuid.UUID) (domain.FlightRoute, error) {
	return s.route, s.err
}
func (s *dbStub) ListSubscribers(context.Context, uuid.UUID) ([]userDomain.User, error) {
	return s.subs, s.err
}
func (s *dbStub) GetFlightAirports(context.Context, uuid.UUID) (airportDomain.Airport, airportDomain.Airport, error) {
	return s.dep, s.arr, s.err
}

type cacheStub struct {
	cachePort
	flight                          domain.Flight
	flights                         []domain.Flight
	err                             error
	saveCalls, getCalls, flushCalls int
}

func (s *cacheStub) SaveFlight(context.Context, domain.Flight) error    { s.saveCalls++; return s.err }
func (s *cacheStub) SaveFlights(context.Context, []domain.Flight) error { s.saveCalls++; return s.err }
func (s *cacheStub) GetFlightById(context.Context, uuid.UUID) (domain.Flight, error) {
	s.getCalls++
	return s.flight, s.err
}
func (s *cacheStub) UpdateFlight(context.Context, domain.UpdateFlightInfo) error { return s.err }
func (s *cacheStub) GetFlights(context.Context) ([]domain.Flight, error) {
	s.getCalls++
	return s.flights, s.err
}
func (s *cacheStub) FlushFlights(context.Context) error { s.flushCalls++; return nil }
func newFixture() (*flightRepository, *dbStub, *cacheStub) {
	db, cache := &dbStub{}, &cacheStub{}
	return &flightRepository{db: db, rd: cache}, db, cache
}

func TestFlightRepository_Save(t *testing.T) {
	t.Run("positive: writes database and cache", func(t *testing.T) {
		// Arrange
		repo, db, cache := newFixture()
		f := domain.Flight{Id: uuid.New()}
		// Act
		err := repo.Save(context.Background(), f)
		// Assert
		if err != nil || db.saveCalls != 1 || cache.saveCalls != 1 {
			t.Fatalf("err=%v, db=%d, cache=%d", err, db.saveCalls, cache.saveCalls)
		}
	})
	t.Run("negative: database failure prevents cache write", func(t *testing.T) {
		// Arrange
		repo, db, cache := newFixture()
		failure := errors.New("database unavailable")
		db.err = failure
		// Act
		err := repo.Save(context.Background(), domain.Flight{})
		// Assert
		if !errors.Is(err, failure) || cache.saveCalls != 0 {
			t.Fatalf("err=%v, cache=%d", err, cache.saveCalls)
		}
	})
}

func TestFlightRepository_Exist(t *testing.T) {
	t.Run("positive: retrieves persisted flight", func(t *testing.T) {
		// Arrange
		repo, db, cache := newFixture()
		fid := uuid.New()
		db.flight = domain.Flight{Id: fid}
		// Act
		got, err := repo.Exist(context.Background(), fid)
		// Assert
		if err != nil || got.Id != fid || db.existCalls != 1 || cache.getCalls != 1 {
			t.Fatalf("got=%+v, err=%v, db=%d, cache=%d", got, err, db.existCalls, cache.getCalls)
		}
	})
	t.Run("negative: missing flight is mapped", func(t *testing.T) {
		// Arrange
		repo, db, _ := newFixture()
		db.err = postgres.ErrFlightNotFound
		// Act
		_, err := repo.Exist(context.Background(), uuid.New())
		// Assert
		if !errors.Is(err, flightDomainRepository.ErrFlightNotFound) {
			t.Fatal(err)
		}
	})
}

func TestFlightRepository_Update(t *testing.T) {
	t.Run("positive: updates database and cache", func(t *testing.T) {
		// Arrange
		repo, db, _ := newFixture()
		// Act
		err := repo.Update(context.Background(), domain.UpdateFlightInfo{FlightId: uuid.New()})
		// Assert
		if err != nil || db.updateCalls != 1 {
			t.Fatalf("err=%v, calls=%d", err, db.updateCalls)
		}
	})
	t.Run("negative: missing flight", func(t *testing.T) {
		// Arrange
		repo, db, _ := newFixture()
		db.err = postgres.ErrFlightNotFound
		// Act
		err := repo.Update(context.Background(), domain.UpdateFlightInfo{FlightId: uuid.New()})
		// Assert
		if !errors.Is(err, flightDomainRepository.ErrFlightNotFound) {
			t.Fatal(err)
		}
	})
}

func TestFlightRepository_ListFlights(t *testing.T) {
	t.Run("positive classic: returns cached flights", func(t *testing.T) {
		// Arrange
		repo, db, cache := newFixture()
		fid := uuid.New()
		cache.flights = []domain.Flight{{Id: fid}}
		// Act
		got, err := repo.ListFlights(context.Background())
		// Assert
		if err != nil || len(got) != 1 || got[0].Id != fid || db.listCalls != 0 {
			t.Fatalf("got=%+v, err=%v, db=%d", got, err, db.listCalls)
		}
	})
	t.Run("negative: database fallback fails", func(t *testing.T) {
		// Arrange
		repo, db, cache := newFixture()
		cache.err = redis.ErrCacheEmpty
		failure := errors.New("database unavailable")
		db.err = failure
		// Act
		_, err := repo.ListFlights(context.Background())
		// Assert
		if !errors.Is(err, failure) || db.listCalls != 1 {
			t.Fatalf("err=%v, db=%d", err, db.listCalls)
		}
	})
}

func TestFlightRepository_GetFlightRoute(t *testing.T) {
	t.Run("positive: returns route", func(t *testing.T) {
		// Arrange
		repo, db, _ := newFixture()
		rid := uuid.New()
		db.route = domain.FlightRoute{Id: rid}
		// Act
		got, err := repo.GetFlightRoute(context.Background(), uuid.New())
		// Assert
		if err != nil || got.Id != rid {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})
	t.Run("negative: missing route is mapped", func(t *testing.T) {
		// Arrange
		repo, db, _ := newFixture()
		db.err = postgres.ErrFlightRouteNotFound
		// Act
		_, err := repo.GetFlightRoute(context.Background(), uuid.New())
		// Assert
		if !errors.Is(err, flightDomainRepository.ErrFlightRouteNotFound) {
			t.Fatal(err)
		}
	})
}

func TestFlightRepository_ListSubscribers(t *testing.T) {
	t.Run("positive: returns subscribers", func(t *testing.T) {
		// Arrange
		repo, db, _ := newFixture()
		uid := uuid.New()
		db.subs = []userDomain.User{{Id: uid}}
		// Act
		got, err := repo.ListSubscribers(context.Background(), uuid.New())
		// Assert
		if err != nil || len(got) != 1 || got[0].Id != uid {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})
	t.Run("negative: database error", func(t *testing.T) {
		// Arrange
		repo, db, _ := newFixture()
		failure := errors.New("database unavailable")
		db.err = failure
		// Act
		_, err := repo.ListSubscribers(context.Background(), uuid.New())
		// Assert
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

func TestFlightRepository_GetFlightAirports(t *testing.T) {
	t.Run("positive: returns both airports", func(t *testing.T) {
		// Arrange
		repo, db, _ := newFixture()
		db.dep = airportDomain.Airport{ID: uuid.New()}
		db.arr = airportDomain.Airport{ID: uuid.New()}
		// Act
		dep, arr, err := repo.GetFlightAirports(context.Background(), uuid.New())
		// Assert
		if err != nil || dep.ID != db.dep.ID || arr.ID != db.arr.ID {
			t.Fatalf("dep=%+v, arr=%+v, err=%v", dep, arr, err)
		}
	})
	t.Run("negative: database error", func(t *testing.T) {
		// Arrange
		repo, db, _ := newFixture()
		failure := errors.New("database unavailable")
		db.err = failure
		// Act
		_, _, err := repo.GetFlightAirports(context.Background(), uuid.New())
		// Assert
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

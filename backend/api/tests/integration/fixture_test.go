//go:build integration

package integration

import (
	"api/internal/flight"
	"api/internal/flight/domain"
	flightPostgres "api/internal/flight/infra/postgres"
	flightRedis "api/internal/flight/infra/redis"
	"api/internal/flight/usecase"
	"api/internal/flight/usecase/repository"
	"api/internal/publisher"
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
)

var seedFlightID = uuid.MustParse("00000000-0000-4000-8000-000000000050")
var adminID = uuid.MustParse("00000000-0000-4000-8000-000000000001")

type fixture struct {
	pool   *pgxpool.Pool
	db     *flightPostgres.PostgresClient
	cache  *flightRedis.RedisClient
	repo   repository.FlightDB
	uc     *usecase.FlightUsecase
	flight domain.Flight
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	if os.Getenv("POSTGRES_URI") == "" || os.Getenv("REDIS_ADDR") == "" {
		t.Fatal("POSTGRES_URI and REDIS_ADDR are required; use make test-integration")
	}
	pool, err := pgxpool.New(t.Context(), os.Getenv("POSTGRES_URI"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}

	client := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})
	t.Cleanup(func() { _ = client.Close() })
	if err := client.Ping(t.Context()).Err(); err != nil {
		t.Fatal(err)
	}

	injector := do.New(flight.Package, publisher.OutboxPackage)
	do.ProvideValue(injector, pool)
	do.ProvideValue(injector, client)
	f := fixture{
		pool:  pool,
		db:    do.MustInvoke[*flightPostgres.PostgresClient](injector),
		cache: do.MustInvoke[*flightRedis.RedisClient](injector),
		repo:  do.MustInvoke[repository.FlightDB](injector),
		uc:    do.MustInvoke[*usecase.FlightUsecase](injector),
	}

	f.flight, err = f.db.Exist(t.Context(), seedFlightID)
	if err != nil {
		t.Fatalf("read fixture flight: %v", err)
	}
	f.flight.Id = uuid.New()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, query := range []string{
			"DELETE FROM outbox WHERE payload->>'flight_id' = $1",
			"DELETE FROM subscriptions WHERE flight_id = $1::uuid",
			"DELETE FROM flight_routes WHERE flight_id = $1::uuid",
			"DELETE FROM flights WHERE id = $1::uuid",
		} {
			if _, err := pool.Exec(ctx, query, f.flight.Id.String()); err != nil {
				t.Errorf("cleanup: %v", err)
			}
		}
		if err := f.cache.DeleteFlightById(ctx, f.flight.Id); err != nil {
			t.Errorf("cleanup cache: %v", err)
		}
	})
	return f
}

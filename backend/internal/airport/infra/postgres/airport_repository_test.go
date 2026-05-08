package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"airline-tracker/internal/airport/domain"
	"airline-tracker/internal/airport/domain/repository"
	"airline-tracker/internal/common"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	connStr := "postgres://" +
		getEnv("DB_USER", "postgres") +
		":" + getEnv("DB_PASSWORD", "postgres") +
		"@" + getEnv("DB_HOST", "localhost") +
		":" + getEnv("DB_PORT", "5432") +
		"/" + getEnv("DB_NAME", "airline_tracker")

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		t.Fatalf("unable to parse postgres config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Skipf("postgres is not available: %v", err)
		return nil
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Skipf("postgres is not available: %v", err)
		return nil
	}

	return pool
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func cleanupAirportByIATA(t *testing.T, pool *pgxpool.Pool, iata string) {
	t.Helper()

	query := `
	delete from airports
		where iata_code = $1
	`

	if _, err := pool.Exec(context.Background(), query, iata); err != nil {
		t.Fatalf("cleanup airport %s: %v", iata, err)
	}
}

func mustNewAirport(t *testing.T, iata, title, city, country string) domain.Airport {
	t.Helper()

	iataCode, err := domain.NewIATACode(iata)
	if err != nil {
		t.Fatalf("invalid iata code %s: %v", iata, err)
	}

	titleValue, err := common.NewTitle(title)
	if err != nil {
		t.Fatalf("invalid title %s: %v", title, err)
	}

	cityValue, err := common.NewCity(city)
	if err != nil {
		t.Fatalf("invalid city %s: %v", city, err)
	}

	countryValue, err := common.NewCountry(country)
	if err != nil {
		t.Fatalf("invalid country %s: %v", country, err)
	}

	airport, err := domain.NewAirport(iataCode, titleValue, cityValue, countryValue)
	if err != nil {
		t.Fatalf("cannot create airport: %v", err)
	}

	return airport
}

func TestAirportRepository_Save(t *testing.T) {
	pool := newTestPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	tests := []struct {
		name     string
		airports []domain.Airport
		cleanup  []string
		verify   func(t *testing.T, pool *pgxpool.Pool, airports []domain.Airport)
	}{
		{
			name: "save one airport",
			airports: []domain.Airport{
				mustNewAirport(t, "SVO", "Sheremetyevo", "Moscow", "Russia"),
			},
			cleanup: []string{"SVO"},
			verify: func(t *testing.T, pool *pgxpool.Pool, airports []domain.Airport) {
				query := `
				select iata_code, title, city, country
					from airports where id = $1
				`
				airport := airports[0]
				var iata, title, city, country string
				row := pool.QueryRow(context.Background(), query, airport.ID)
				if err := row.Scan(&iata, &title, &city, &country); err != nil {
					t.Fatalf("expected airport row, got: %v", err)
				}

				if iata != string(airport.IATACode) ||
					title != string(airport.Title) ||
					city != string(airport.City) ||
					country != string(airport.Country) {
					t.Fatalf("expected: %s %s %s %s\ngot: %s %s %s %s",
						airport.IATACode, airport.Title, airport.City, airport.Country,
						iata, title, city, country,
					)
				}
			},
		},
		{
			name: "save multiple airports",
			airports: []domain.Airport{
				mustNewAirport(t, "JFK", "John F. Kennedy", "New York", "USA"),
				mustNewAirport(t, "LHR", "Heathrow", "London", "UK"),
			},
			cleanup: []string{"JFK", "LHR"},
			verify: func(t *testing.T, pool *pgxpool.Pool, airports []domain.Airport) {
				query := `
				select count(*) from airports
					where iata_code in ($1, $2)
				`
				var count int
				row := pool.QueryRow(context.Background(), query,
					airports[0].IATACode, airports[1].IATACode)
				if err := row.Scan(&count); err != nil {
					t.Fatalf("query failed, got: %v", err)
				}
				if count != 2 {
					t.Fatalf("expected 2 airports, got %d", count)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &airportRepository{conn: pool}
			for _, airport := range tt.airports {
				if err := repo.Save(context.Background(), airport); err != nil {
					t.Fatalf("error while saving airport, got: %v", err)
				}
			}

			tt.verify(t, pool, tt.airports)

			for _, iata := range tt.cleanup {
				cleanupAirportByIATA(t, pool, iata)
			}
		})
	}
}

func TestAirportRepository_Save_DuplicateIATA(t *testing.T) {
	pool := newTestPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	tests := []struct {
		name      string
		first     domain.Airport
		duplicate domain.Airport
		cleanup   []string
		expected  error
	}{
		{
			name:      "duplicate iata code",
			first:     mustNewAirport(t, "CDG", "Charles de Gaulle", "Paris", "France"),
			duplicate: mustNewAirport(t, "CDG", "Paris North", "Paris", "France"),
			cleanup:   []string{"CDG"},
			expected:  repository.ErrAirportAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, iata := range tt.cleanup {
				cleanupAirportByIATA(t, pool, iata)
				defer cleanupAirportByIATA(t, pool, iata)
			}

			repo := &airportRepository{conn: pool}
			if err := repo.Save(context.Background(), tt.first); err != nil {
				t.Fatalf("error while saving first airport, got: %v", err)
			}

			if err := repo.Save(context.Background(), tt.duplicate); err == nil {
				t.Fatal("saved second airport")
			} else if !errors.Is(err, tt.expected) {
				t.Fatalf("expected %v, got: %v", tt.expected, err)
			}
		})
	}
}

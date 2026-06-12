package usecase

import (
	"api/internal/airport/command"
	"api/internal/airport/domain"
	"api/internal/airport/domain/repository"
	"api/internal/airport/query"
	"context"
	"errors"
	"testing"

	"github.com/samber/do/v2"
)

type airportRepoMock struct {
	saveFn func(ctx context.Context, a domain.Airport) error

	listAirportsFn func(ctx context.Context) ([]domain.Airport, error)
}

func (m *airportRepoMock) Save(ctx context.Context, a domain.Airport) error {
	return m.saveFn(ctx, a)
}

func (m *airportRepoMock) ListAirports(ctx context.Context) ([]domain.Airport, error) {
	return m.listAirportsFn(ctx)
}

func TestAirportUsecase_CreateAirport(t *testing.T) {
	tests := []struct {
		name    string
		repo    repository.AirportRepository
		cmd     *command.CreateAirportCommand
		wantErr bool
	}{
		{
			name: "[P] valid airport",
			repo: &airportRepoMock{
				saveFn: func(ctx context.Context, a domain.Airport) error {
					return nil
				},
			},
			cmd: &command.CreateAirportCommand{
				IATACode: "SVO",
				Title:    "Domodedovo",
				City:     "Moscow",
				Country:  "RU",
			},
			wantErr: false,
		},

		{
			name: "[N:IATA] invalid",
			repo: &airportRepoMock{
				saveFn: func(ctx context.Context, a domain.Airport) error {
					return nil
				},
			},
			cmd: &command.CreateAirportCommand{
				IATACode: "",
				Title:    "Domodedovo",
				City:     "Moscow",
				Country:  "RU",
			},
			wantErr: true,
		},

		{
			name: "[N:Title] invalid",
			repo: &airportRepoMock{
				saveFn: func(ctx context.Context, a domain.Airport) error {
					return nil
				},
			},
			cmd: &command.CreateAirportCommand{
				IATACode: "SVO",
				Title:    "",
				City:     "Moscow",
				Country:  "RU",
			},
			wantErr: true,
		},

		{
			name: "[N:City] invalid",
			repo: &airportRepoMock{
				saveFn: func(ctx context.Context, a domain.Airport) error {
					return nil
				},
			},
			cmd: &command.CreateAirportCommand{
				IATACode: "SVO",
				Title:    "Domodedovo",
				City:     "",
				Country:  "RU",
			},
			wantErr: true,
		},

		{
			name: "[N:Country] invalid",
			repo: &airportRepoMock{
				saveFn: func(ctx context.Context, a domain.Airport) error {
					return nil
				},
			},
			cmd: &command.CreateAirportCommand{
				IATACode: "SVO",
				Title:    "Domodedovo",
				City:     "Moscow",
				Country:  "",
			},
			wantErr: true,
		},

		{
			name: "[N] airport already exists",
			repo: &airportRepoMock{
				saveFn: func(ctx context.Context, a domain.Airport) error {
					return repository.ErrAirportAlreadyExists
				},
			},
			cmd: &command.CreateAirportCommand{
				IATACode: "SVO",
				Title:    "Domodedovo",
				City:     "Moscow",
				Country:  "RU",
			},
			wantErr: true,
		},
		{
			name: "[N] repo error",
			repo: &airportRepoMock{
				saveFn: func(ctx context.Context, a domain.Airport) error {
					return errors.New("repo error")
				},
			},
			cmd: &command.CreateAirportCommand{
				IATACode: "SVO",
				Title:    "Domodedovo",
				City:     "Moscow",
				Country:  "RU",
			},
			wantErr: true,
		},
	}

	injector := do.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			do.Override(injector, func(i do.Injector) (repository.AirportRepository, error) {
				return tt.repo, nil
			})
			uc, err := NewAirportUsecase(injector)
			if err != nil {
				t.Fatalf("NewAirportUsecase: err = %v", err)
			}
			err = uc.CreateAirport(tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestAirportUsecase_ListAirports(t *testing.T) {
	tests := []struct {
		name    string
		repo    repository.AirportRepository
		want    query.ListAirportsQuery
		wantErr bool
	}{
		{
			name: "[P] empty result",
			repo: &airportRepoMock{
				listAirportsFn: func(ctx context.Context) ([]domain.Airport, error) {
					return nil, nil
				},
			},
			want:    query.ListAirportsQuery{},
			wantErr: false,
		},
		{
			name: "[N] repo error",
			repo: &airportRepoMock{
				listAirportsFn: func(ctx context.Context) ([]domain.Airport, error) {
					return nil, errors.New("repo error")
				},
			},
			want:    query.ListAirportsQuery{},
			wantErr: true,
		},
	}
	injector := do.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			do.Override(injector, func(i do.Injector) (repository.AirportRepository, error) {
				return tt.repo, nil
			})
			uc, err := NewAirportUsecase(injector)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got, gotErr := uc.ListAirports()
			if gotErr != nil {
				if !tt.wantErr {
					t.Fatalf("gotErr = %v, wantErr = %v", err, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListAirports() succeeded unexpectedly")
			}

			if len(got.Airports) != len(tt.want.Airports) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

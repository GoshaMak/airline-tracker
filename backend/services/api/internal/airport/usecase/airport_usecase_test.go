package usecase

import (
	"api/internal/airport/command"
	"api/internal/airport/domain"
	"api/internal/airport/domain/repository"
	"api/internal/airport/dto"
	"context"
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

func TestAirportUsecase(t *testing.T) {
	tests := []struct {
		name    string
		repo    *airportRepoMock
		data    any
		wantErr bool
	}{
		{
			name: "[IATA code] empty",
			repo: &airportRepoMock{
				saveFn: func(ctx context.Context, a domain.Airport) error {
					return nil
				},
			},
			data: &dto.CreateAirportRequest{
				Airport: dto.AirportDTO{
					IATACode: "",
					Title:    "Domodedovo",
					City:     "Moscow",
					Country:  "RU",
				},
			},
			wantErr: true,
		},
		{
			name: "[IATA code] wrong format",
			repo: &airportRepoMock{
				saveFn: func(ctx context.Context, a domain.Airport) error {
					return nil
				},
			},
			data: &dto.CreateAirportRequest{
				Airport: dto.AirportDTO{
					IATACode: "Domodedovo",
					Title:    "Domodedovo",
					City:     "Moscow",
					Country:  "RU",
				},
			},
			wantErr: true,
		},
		{
			name: "[IATA code] correct format",
			repo: &airportRepoMock{
				saveFn: func(ctx context.Context, a domain.Airport) error {
					return nil
				},
			},
			data: &dto.CreateAirportRequest{
				Airport: dto.AirportDTO{
					IATACode: "SVO",
					Title:    "Domodedovo",
					City:     "Moscow",
					Country:  "RU",
				},
			},
			wantErr: false,
		},
	}
	injector := do.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			do.Override(injector, func(i do.Injector) (repository.AirportRepository, error) {
				return tt.repo, nil
			})
			s, _ := NewAirportUsecase(injector)
			req := tt.data.(*dto.CreateAirportRequest)
			cmd := &command.CreateAirportCommand{
				IATACode: req.Airport.IATACode,
				Title:    req.Airport.Title,
				City:     req.Airport.City,
				Country:  req.Airport.Country,
			}
			err := s.CreateAirport(cmd)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

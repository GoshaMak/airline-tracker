package usecase

import (
	"airline-tracker/internal/airport/command"
	"airline-tracker/internal/airport/domain"
	"airline-tracker/internal/airport/domain/repository"
	"airline-tracker/internal/airport/dto"
	"context"
	"testing"

	"github.com/samber/do/v2"
)

type airportRepoMock struct {
	saveFn func(ctx context.Context, a *domain.Airport) error
}

func (m *airportRepoMock) Save(ctx context.Context, a *domain.Airport) error {
	return m.saveFn(ctx, a)
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
				saveFn: func(ctx context.Context, a *domain.Airport) error {
					return nil
				},
			},
			data: &dto.CreateAirportRequest{
				Airport: dto.AirportDTO{
					IATACode: "",
					Title:    "Title",
					City:     "City",
					Country:  "Country",
				},
			},
			wantErr: true,
		},
		{
			name: "[IATA code] wrong format",
			repo: &airportRepoMock{
				saveFn: func(ctx context.Context, a *domain.Airport) error {
					return nil
				},
			},
			data: &dto.CreateAirportRequest{
				Airport: dto.AirportDTO{
					IATACode: "123",
					Title:    "Title",
					City:     "City",
					Country:  "Country",
				},
			},
			wantErr: true,
		},
		{
			name: "[IATA code] correct format",
			repo: &airportRepoMock{
				saveFn: func(ctx context.Context, a *domain.Airport) error {
					return nil
				},
			},
			data: &dto.CreateAirportRequest{
				Airport: dto.AirportDTO{
					IATACode: "SVO",
					Title:    "Title",
					City:     "City",
					Country:  "Country",
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

package usecase

import (
	"api/internal/fleet/command"
	"api/internal/fleet/domain"
	"api/internal/fleet/domain/repository"
	"context"
	"errors"
	"testing"

	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type aircraftRepoMock struct {
	saveFn func(ctx context.Context, a domain.Aircraft) error
	listFn func(ctx context.Context) ([]domain.Aircraft, error)
}

func (m *aircraftRepoMock) SaveAircraft(ctx context.Context, a domain.Aircraft) error {
	return m.saveFn(ctx, a)
}

func (m *aircraftRepoMock) List(ctx context.Context) ([]domain.Aircraft, error) {
	return m.listFn(ctx)
}

func TestAircraftUsecase_CreateAircraft(t *testing.T) {
	tests := []struct {
		name    string
		repo    repository.AircraftRepository
		cmd     command.CreateAircraftCommand
		wantErr bool
	}{
		{
			name: "[P] valid aircraft",
			repo: &aircraftRepoMock{
				saveFn: func(ctx context.Context, a domain.Aircraft) error {
					return nil
				},
			},
			cmd: command.CreateAircraftCommand{
				RegistrationNumber: "N-1234",
				AircraftModelID:    uuid.New(),
				SerialNumber:       "SN-1234",
				Mileage:            100,
			},
			wantErr: false,
		},
		{
			name: "[N] invalid registration number",
			repo: &aircraftRepoMock{
				saveFn: func(ctx context.Context, a domain.Aircraft) error {
					return nil
				},
			},
			cmd: command.CreateAircraftCommand{
				RegistrationNumber: "",
				AircraftModelID:    uuid.New(),
				SerialNumber:       "SN-1234",
				Mileage:            100,
			},
			wantErr: true,
		},
		{
			name: "[N] invalid serial number",
			repo: &aircraftRepoMock{
				saveFn: func(ctx context.Context, a domain.Aircraft) error {
					return nil
				},
			},
			cmd: command.CreateAircraftCommand{
				RegistrationNumber: "N-1234",
				AircraftModelID:    uuid.New(),
				SerialNumber:       "",
				Mileage:            100,
			},
			wantErr: true,
		},
		{
			name: "[N] invalid mileage",
			repo: &aircraftRepoMock{
				saveFn: func(ctx context.Context, a domain.Aircraft) error {
					return nil
				},
			},
			cmd: command.CreateAircraftCommand{
				RegistrationNumber: "N-1234",
				AircraftModelID:    uuid.New(),
				SerialNumber:       "SN-1234",
				Mileage:            -1,
			},
			wantErr: true,
		},
		{
			name: "[N] already exists",
			repo: &aircraftRepoMock{
				saveFn: func(ctx context.Context, a domain.Aircraft) error {
					return repository.ErrAircraftAlreadyExists
				},
			},
			cmd: command.CreateAircraftCommand{
				RegistrationNumber: "N-1234",
				AircraftModelID:    uuid.New(),
				SerialNumber:       "SN-1234",
				Mileage:            100,
			},
			wantErr: true,
		},
		{
			name: "[N] repo error",
			repo: &aircraftRepoMock{
				saveFn: func(ctx context.Context, a domain.Aircraft) error {
					return errors.New("repo error")
				},
			},
			cmd: command.CreateAircraftCommand{
				RegistrationNumber: "N-1234",
				AircraftModelID:    uuid.New(),
				SerialNumber:       "SN-1234",
				Mileage:            100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		allure.Test(t, tt.name, func(allureContext *allure.Context) {
			t := allureContext.T()

			injector := do.New()
			do.Override(injector, func(i do.Injector) (repository.AircraftRepository, error) {
				return tt.repo, nil
			})

			uc, err := NewAircraftUsecase(injector)
			if err != nil {
				t.Fatalf("NewAircraftUsecase: err = %v", err)
			}

			gotErr := uc.CreateAircraft(tt.cmd)
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestAircraftUsecase_ListAircrafts(t *testing.T) {
	tests := []struct {
		name    string
		repo    repository.AircraftRepository
		wantLen int
		wantErr bool
	}{
		{
			name: "[P] empty list",
			repo: &aircraftRepoMock{
				listFn: func(ctx context.Context) ([]domain.Aircraft, error) {
					return nil, nil
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "[P] non-empty list",
			repo: &aircraftRepoMock{
				listFn: func(ctx context.Context) ([]domain.Aircraft, error) {
					return []domain.Aircraft{{}, {}}, nil
				},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "[N] repo error",
			repo: &aircraftRepoMock{
				listFn: func(ctx context.Context) ([]domain.Aircraft, error) {
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
			do.Override(injector, func(i do.Injector) (repository.AircraftRepository, error) {
				return tt.repo, nil
			})

			uc, err := NewAircraftUsecase(injector)
			if err != nil {
				t.Fatalf("NewAircraftUsecase: err = %v", err)
			}

			got, gotErr := uc.ListAircrafts()
			if gotErr != nil {
				if !tt.wantErr {
					t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListAircrafts() succeeded unexpectedly")
			}
			if len(got) != tt.wantLen {
				t.Fatalf("got len = %d, want = %d", len(got), tt.wantLen)
			}
		})
	}
}

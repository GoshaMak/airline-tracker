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

type aircraftModelRepoMock struct {
	saveFn func(ctx context.Context, am domain.AircraftModel) error
	getFn  func(ctx context.Context, id uuid.UUID) (domain.AircraftModel, error)
}

func (m *aircraftModelRepoMock) SaveAircraftModel(ctx context.Context, am domain.AircraftModel) error {
	return m.saveFn(ctx, am)
}

func (m *aircraftModelRepoMock) GetAircraftModelById(ctx context.Context, id uuid.UUID) (domain.AircraftModel, error) {
	return m.getFn(ctx, id)
}

func TestAircraftModelUsecase_CreateAircraftModel(t *testing.T) {
	tests := []struct {
		name    string
		repo    repository.AircraftModelRepository
		cmd     command.CreateAircraftModelCommand
		wantErr bool
	}{
		{
			name: "[P] valid aircraft model",
			repo: &aircraftModelRepoMock{
				saveFn: func(ctx context.Context, am domain.AircraftModel) error {
					return nil
				},
			},
			cmd: command.CreateAircraftModelCommand{
				Manufacturer: "Boeing",
				Model:        "737 MAX",
				Mass:         10000,
				MaxAltitude:  12000,
				MaxSpeed:     850,
			},
			wantErr: false,
		},
		{
			name: "[N] invalid manufacturer",
			repo: &aircraftModelRepoMock{
				saveFn: func(ctx context.Context, am domain.AircraftModel) error {
					return nil
				},
			},
			cmd: command.CreateAircraftModelCommand{
				Manufacturer: "",
				Model:        "737 MAX",
				Mass:         10000,
				MaxAltitude:  12000,
				MaxSpeed:     850,
			},
			wantErr: true,
		},
		{
			name: "[N] invalid model",
			repo: &aircraftModelRepoMock{
				saveFn: func(ctx context.Context, am domain.AircraftModel) error {
					return nil
				},
			},
			cmd: command.CreateAircraftModelCommand{
				Manufacturer: "Boeing",
				Model:        "",
				Mass:         10000,
				MaxAltitude:  12000,
				MaxSpeed:     850,
			},
			wantErr: true,
		},
		{
			name: "[N] invalid mass",
			repo: &aircraftModelRepoMock{
				saveFn: func(ctx context.Context, am domain.AircraftModel) error {
					return nil
				},
			},
			cmd: command.CreateAircraftModelCommand{
				Manufacturer: "Boeing",
				Model:        "737 MAX",
				Mass:         -1,
				MaxAltitude:  12000,
				MaxSpeed:     850,
			},
			wantErr: true,
		},
		{
			name: "[N] invalid max altitude",
			repo: &aircraftModelRepoMock{
				saveFn: func(ctx context.Context, am domain.AircraftModel) error {
					return nil
				},
			},
			cmd: command.CreateAircraftModelCommand{
				Manufacturer: "Boeing",
				Model:        "737 MAX",
				Mass:         10000,
				MaxAltitude:  -1,
				MaxSpeed:     850,
			},
			wantErr: true,
		},
		{
			name: "[N] invalid max speed",
			repo: &aircraftModelRepoMock{
				saveFn: func(ctx context.Context, am domain.AircraftModel) error {
					return nil
				},
			},
			cmd: command.CreateAircraftModelCommand{
				Manufacturer: "Boeing",
				Model:        "737 MAX",
				Mass:         10000,
				MaxAltitude:  12000,
				MaxSpeed:     -1,
			},
			wantErr: true,
		},
		{
			name: "[N] already exists",
			repo: &aircraftModelRepoMock{
				saveFn: func(ctx context.Context, am domain.AircraftModel) error {
					return repository.ErrAircraftModelAlreadyExists
				},
			},
			cmd: command.CreateAircraftModelCommand{
				Manufacturer: "Boeing",
				Model:        "737 MAX",
				Mass:         10000,
				MaxAltitude:  12000,
				MaxSpeed:     850,
			},
			wantErr: true,
		},
		{
			name: "[N] repo error",
			repo: &aircraftModelRepoMock{
				saveFn: func(ctx context.Context, am domain.AircraftModel) error {
					return errors.New("repo error")
				},
			},
			cmd: command.CreateAircraftModelCommand{
				Manufacturer: "Boeing",
				Model:        "737 MAX",
				Mass:         10000,
				MaxAltitude:  12000,
				MaxSpeed:     850,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		allure.Test(t, tt.name, func(allureContext *allure.Context) {
			t := allureContext.T()

			injector := do.New()
			do.Override(injector, func(i do.Injector) (repository.AircraftModelRepository, error) {
				return tt.repo, nil
			})

			uc, err := NewAircraftModelUsecase(injector)
			if err != nil {
				t.Fatalf("NewAircraftModelUsecase: err = %v", err)
			}

			gotErr := uc.CreateAircraftModel(tt.cmd)
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestAircraftModelUsecase_AircraftById(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name    string
		repo    repository.AircraftModelRepository
		wantErr bool
	}{
		{
			name: "[P] found",
			repo: &aircraftModelRepoMock{
				getFn: func(ctx context.Context, id uuid.UUID) (domain.AircraftModel, error) {
					return domain.AircraftModel{Id: id}, nil
				},
			},
			wantErr: false,
		},
		{
			name: "[N] not found",
			repo: &aircraftModelRepoMock{
				getFn: func(ctx context.Context, id uuid.UUID) (domain.AircraftModel, error) {
					return domain.AircraftModel{}, repository.ErrAircraftModelNotFound
				},
			},
			wantErr: true,
		},
		{
			name: "[N] repo error",
			repo: &aircraftModelRepoMock{
				getFn: func(ctx context.Context, id uuid.UUID) (domain.AircraftModel, error) {
					return domain.AircraftModel{}, errors.New("repo error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		allure.Test(t, tt.name, func(allureContext *allure.Context) {
			t := allureContext.T()

			injector := do.New()
			do.Override(injector, func(i do.Injector) (repository.AircraftModelRepository, error) {
				return tt.repo, nil
			})

			uc, err := NewAircraftModelUsecase(injector)
			if err != nil {
				t.Fatalf("NewAircraftModelUsecase: err = %v", err)
			}

			_, gotErr := uc.AircraftById(id)
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
			}
		})
	}
}

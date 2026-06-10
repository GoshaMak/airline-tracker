package usecase

import (
	"api/internal/airport/command"
	"api/internal/airport/domain"
	"api/internal/airport/domain/repository"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type gateRepoMock struct {
	saveFn func(ctx context.Context, g domain.Gate) error

	getAirportByGateIdFn func(ctx context.Context, gid uuid.UUID) (domain.Airport, error)

	listFn func(ctx context.Context) ([]domain.Gate, error)
}

func (m *gateRepoMock) Save(ctx context.Context, g domain.Gate) error {
	return m.saveFn(ctx, g)
}

func (m *gateRepoMock) GetAirportByGateId(ctx context.Context, gid uuid.UUID) (domain.Airport, error) {
	return m.getAirportByGateIdFn(ctx, gid)
}

func (m *gateRepoMock) List(ctx context.Context) ([]domain.Gate, error) {
	return m.listFn(ctx)
}

func TestGateUsecase_CreateGate(t *testing.T) {
	validGateNumber := "A1"

	tests := []struct {
		name    string
		repo    repository.GateRepository
		cmd     *command.CreateGateCommand
		wantErr bool
	}{
		{
			name: "[P] valid gate",
			repo: &gateRepoMock{
				saveFn: func(ctx context.Context, g domain.Gate) error {
					return nil
				},
			},
			cmd: &command.CreateGateCommand{
				AirportID:  uuid.New(),
				GateNumber: validGateNumber,
			},
			wantErr: false,
		},
		{
			name: "[N] invalid gate number",
			repo: &gateRepoMock{
				saveFn: func(ctx context.Context, g domain.Gate) error {
					return nil
				},
			},
			cmd: &command.CreateGateCommand{
				AirportID:  uuid.New(),
				GateNumber: "invalid",
			},
			wantErr: true,
		},
		{
			name: "[N] gate already exists",
			repo: &gateRepoMock{
				saveFn: func(ctx context.Context, g domain.Gate) error {
					return repository.ErrGateAlreadyExists
				},
			},
			cmd: &command.CreateGateCommand{
				AirportID:  uuid.New(),
				GateNumber: validGateNumber,
			},
			wantErr: true,
		},
		{
			name: "[N] repo error",
			repo: &gateRepoMock{
				saveFn: func(ctx context.Context, g domain.Gate) error {
					return errors.New("repo error")
				},
			},
			cmd: &command.CreateGateCommand{
				AirportID:  uuid.New(),
				GateNumber: validGateNumber,
			},
			wantErr: true,
		},
	}

	injector := do.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			do.Override(injector, func(i do.Injector) (repository.GateRepository, error) {
				return tt.repo, nil
			})

			uc, gotErr := NewGateUsecase(injector)
			if gotErr != nil {
				t.Fatalf("NewGateUsecase: err = %v", gotErr)
			}

			gotErr = uc.CreateGate(tt.cmd)
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestGateUsecase_ListGates(t *testing.T) {
	tests := []struct {
		name    string
		repo    repository.GateRepository
		wantLen int
		wantErr bool
	}{
		{
			name: "[P] empty list",
			repo: &gateRepoMock{
				listFn: func(ctx context.Context) ([]domain.Gate, error) {
					return nil, nil
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "[P] non-empty list",
			repo: &gateRepoMock{
				listFn: func(ctx context.Context) ([]domain.Gate, error) {
					return []domain.Gate{{}, {}}, nil
				},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "[N] repo error",
			repo: &gateRepoMock{
				listFn: func(ctx context.Context) ([]domain.Gate, error) {
					return nil, errors.New("db error")
				},
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	injector := do.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			do.Override(injector, func(i do.Injector) (repository.GateRepository, error) {
				return tt.repo, nil
			})

			uc, err := NewGateUsecase(injector)
			if err != nil {
				t.Fatalf("NewGateUsecase: err = %v", err)
			}

			got, err := uc.ListGates()
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("ListGates() succeeded unexpectedly")
			}

			if len(got) != tt.wantLen {
				t.Fatalf("got len = %d, want = %d", len(got), tt.wantLen)
			}
		})
	}
}

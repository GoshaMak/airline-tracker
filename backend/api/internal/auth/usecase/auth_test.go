package usecase

import (
	flightDomain "api/internal/flight/domain"
	"api/internal/user/domain"
	"api/internal/user/domain/repository"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type userRepoMock struct {
	saveUserFn func(ctx context.Context, u domain.User) error
	getUserFn  func(ctx context.Context, email string) (domain.User, error)

	existFn     func(ctx context.Context, uid uuid.UUID) (domain.User, error)
	subscribeFn func(ctx context.Context, uid, fid uuid.UUID) error
	listFn      func(ctx context.Context, uid uuid.UUID) ([]flightDomain.Flight, error)
}

func (m *userRepoMock) SaveUser(ctx context.Context, u domain.User) error {
	return m.saveUserFn(ctx, u)
}

func (m *userRepoMock) GetUser(ctx context.Context, email string) (domain.User, error) {
	return m.getUserFn(ctx, email)
}

func (m *userRepoMock) Exist(ctx context.Context, uid uuid.UUID) (domain.User, error) {
	return m.existFn(ctx, uid)
}

func (m *userRepoMock) Subscribe(ctx context.Context, uid, fid uuid.UUID) error {
	return m.subscribeFn(ctx, uid, fid)
}

func (m *userRepoMock) ListFlights(ctx context.Context, uid uuid.UUID) ([]flightDomain.Flight, error) {
	return m.listFn(ctx, uid)
}

func TestAuthUsecase_GetUser(t *testing.T) {
	validUser, err := domain.NewUser("user@example.com", "Aa1!aaaa", domain.UserRole)
	if err != nil {
		t.Fatalf("domain.NewUser: err = %v", err)
	}

	tests := []struct {
		name    string
		repo    repository.UserRepository
		email   string
		pass    string
		wantErr bool
	}{
		{
			name: "[P] valid credentials",
			repo: &userRepoMock{
				getUserFn: func(ctx context.Context, email string) (domain.User, error) {
					return validUser, nil
				},
			},
			email:   validUser.Email.String(),
			pass:    "Aa1!aaaa",
			wantErr: false,
		},
		{
			name: "[N] user not found",
			repo: &userRepoMock{
				getUserFn: func(ctx context.Context, email string) (domain.User, error) {
					return domain.User{}, repository.ErrUserNotFound
				},
			},
			email:   "missing@example.com",
			pass:    "Aa1!aaaa",
			wantErr: true,
		},
		{
			name: "[N] invalid password",
			repo: &userRepoMock{
				getUserFn: func(ctx context.Context, email string) (domain.User, error) {
					return validUser, nil
				},
			},
			email:   validUser.Email.String(),
			pass:    "wrong-pass",
			wantErr: true,
		},
		{
			name: "[N] repo error",
			repo: &userRepoMock{
				getUserFn: func(ctx context.Context, email string) (domain.User, error) {
					return domain.User{}, errors.New("repo error")
				},
			},
			email:   validUser.Email.String(),
			pass:    "Aa1!aaaa",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			injector := do.New()
			do.Override(injector, func(i do.Injector) (repository.UserRepository, error) {
				return tt.repo, nil
			})

			uc, err := NewAuthUsecase(injector)
			if err != nil {
				t.Fatalf("NewAuthUsecase: err = %v", err)
			}

			_, gotErr := uc.GetUser(tt.email, tt.pass)
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestAuthUsecase_CreateUser(t *testing.T) {
	validUser, err := domain.NewUser("user@example.com", "Aa1!aaaa", domain.UserRole)
	if err != nil {
		t.Fatalf("domain.NewUser: err = %v", err)
	}

	tests := []struct {
		name    string
		repo    repository.UserRepository
		wantErr bool
	}{
		{
			name: "[P] valid user",
			repo: &userRepoMock{
				saveUserFn: func(ctx context.Context, u domain.User) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "[N] already exists",
			repo: &userRepoMock{
				saveUserFn: func(ctx context.Context, u domain.User) error {
					return repository.ErrUserAlreadyExists
				},
			},
			wantErr: true,
		},
		{
			name: "[N] repo error",
			repo: &userRepoMock{
				saveUserFn: func(ctx context.Context, u domain.User) error {
					return errors.New("repo error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			injector := do.New()
			do.Override(injector, func(i do.Injector) (repository.UserRepository, error) {
				return tt.repo, nil
			})

			uc, err := NewAuthUsecase(injector)
			if err != nil {
				t.Fatalf("NewAuthUsecase: err = %v", err)
			}

			gotErr := uc.CreateUser(validUser)
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestAuthUsecase_Exists(t *testing.T) {
	validUser, err := domain.NewUser("user@example.com", "Aa1!aaaa", domain.UserRole)
	if err != nil {
		t.Fatalf("domain.NewUser: err = %v", err)
	}

	tests := []struct {
		name string
		repo repository.UserRepository
		want bool
	}{
		{
			name: "[P] exists",
			repo: &userRepoMock{
				getUserFn: func(ctx context.Context, email string) (domain.User, error) {
					return validUser, nil
				},
			},
			want: true,
		},
		{
			name: "[N] missing",
			repo: &userRepoMock{
				getUserFn: func(ctx context.Context, email string) (domain.User, error) {
					return domain.User{}, repository.ErrUserNotFound
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			injector := do.New()
			do.Override(injector, func(i do.Injector) (repository.UserRepository, error) {
				return tt.repo, nil
			})

			uc, err := NewAuthUsecase(injector)
			if err != nil {
				t.Fatalf("NewAuthUsecase: err = %v", err)
			}

			got := uc.Exists(validUser.Email.String(), "Aa1!aaaa")
			if got != tt.want {
				t.Fatalf("got = %v, want = %v", got, tt.want)
			}
		})
	}
}

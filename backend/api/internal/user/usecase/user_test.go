package usecase

import (
	airportDomain "api/internal/airport/domain"
	airportRepository "api/internal/airport/domain/repository"
	flightDomain "api/internal/flight/domain"
	flightRepository "api/internal/flight/domain/repository"
	publisherDomain "api/internal/publisher/domain"
	publisherRepository "api/internal/publisher/domain/repository"
	userDomain "api/internal/user/domain"
	userRepository "api/internal/user/domain/repository"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"shared/common"
)

type userRepoMock struct {
	saveUserFn  func(ctx context.Context, u userDomain.User) error
	getUserFn   func(ctx context.Context, email string) (userDomain.User, error)
	existFn     func(ctx context.Context, uid uuid.UUID) (userDomain.User, error)
	subscribeFn func(ctx context.Context, uid, fid uuid.UUID) error
	listFn      func(ctx context.Context, uid uuid.UUID) ([]flightDomain.Flight, error)
}

func (m *userRepoMock) SaveUser(ctx context.Context, u userDomain.User) error {
	return m.saveUserFn(ctx, u)
}

func (m *userRepoMock) GetUser(ctx context.Context, email string) (userDomain.User, error) {
	return m.getUserFn(ctx, email)
}

func (m *userRepoMock) Exist(ctx context.Context, uid uuid.UUID) (userDomain.User, error) {
	return m.existFn(ctx, uid)
}

func (m *userRepoMock) Subscribe(ctx context.Context, uid, fid uuid.UUID) error {
	return m.subscribeFn(ctx, uid, fid)
}

func (m *userRepoMock) ListFlights(ctx context.Context, uid uuid.UUID) ([]flightDomain.Flight, error) {
	return m.listFn(ctx, uid)
}

type flightRepoMock struct {
	existFn func(ctx context.Context, fid uuid.UUID) (flightDomain.Flight, error)
	routeFn func(ctx context.Context, fid uuid.UUID) (flightDomain.FlightRoute, error)
	listFn  func(ctx context.Context, fid uuid.UUID) ([]userDomain.User, error)
}

func (m *flightRepoMock) Save(ctx context.Context, flight flightDomain.Flight) error { return nil }

func (m *flightRepoMock) Exist(ctx context.Context, fid uuid.UUID) (flightDomain.Flight, error) {
	return m.existFn(ctx, fid)
}

func (m *flightRepoMock) Update(ctx context.Context, ufi flightDomain.UpdateFlightInfo) error {
	return nil
}

func (m *flightRepoMock) ListFlights(ctx context.Context) ([]flightDomain.Flight, error) {
	return nil, nil
}

func (m *flightRepoMock) GetFlightRoute(ctx context.Context, fid uuid.UUID) (flightDomain.FlightRoute, error) {
	return m.routeFn(ctx, fid)
}

func (m *flightRepoMock) ListSubscribers(ctx context.Context, fid uuid.UUID) ([]userDomain.User, error) {
	return m.listFn(ctx, fid)
}

func (m *flightRepoMock) GetFlightAirports(
	ctx context.Context,
	fid uuid.UUID,
) (airportDomain.Airport, airportDomain.Airport, error) {
	return airportDomain.Airport{}, airportDomain.Airport{}, nil
}

type gateRepoMock struct {
	airportFn func(ctx context.Context, gid uuid.UUID) (airportDomain.Airport, error)
}

func (m *gateRepoMock) Save(ctx context.Context, g airportDomain.Gate) error { return nil }

func (m *gateRepoMock) GetAirportByGateId(ctx context.Context, gid uuid.UUID) (airportDomain.Airport, error) {
	return m.airportFn(ctx, gid)
}

func (m *gateRepoMock) List(ctx context.Context) ([]airportDomain.Gate, error) { return nil, nil }

type outboxRepoMock struct {
	saveFn func(ctx context.Context, ob publisherDomain.Outbox) error
}

func (m *outboxRepoMock) Save(ctx context.Context, ob publisherDomain.Outbox) error {
	return m.saveFn(ctx, ob)
}

func (m *outboxRepoMock) ListNotSent(
	ctx context.Context,
	newPayload func() publisherDomain.Payload,
) ([]publisherDomain.Outbox, error) {
	return nil, nil
}

func (m *outboxRepoMock) MarkAsSent(ctx context.Context, ob publisherDomain.Outbox) error {
	return nil
}

func validUser(t *testing.T) userDomain.User {
	t.Helper()
	u, err := userDomain.NewUser("user@example.com", "Aa1!aaaa", userDomain.UserRole)
	if err != nil {
		t.Fatalf("domain.NewUser: err = %v", err)
	}
	return u
}

func TestUserUsecase_Exist(t *testing.T) {
	// Arrange: the existing repository mock provides the user used by both cases.
	u := validUser(t)
	repo := &userRepoMock{getUserFn: func(_ context.Context, email string) (userDomain.User, error) {
		if email == u.Email.String() {
			return u, nil
		}
		return userDomain.User{}, userRepository.ErrUserNotFound
	}}
	uc := &UserUsecase{userRepo: repo}

	t.Run("positive: correct password", func(t *testing.T) {
		// Act
		got := uc.Exist(u.Email.String(), "Aa1!aaaa")
		// Assert
		if !got {
			t.Fatal("existing user was rejected")
		}
	})
	t.Run("negative: wrong password", func(t *testing.T) {
		// Act
		got := uc.Exist(u.Email.String(), "wrong-password")
		// Assert
		if got {
			t.Fatal("invalid credentials were accepted")
		}
	})
}

func validAirport(t *testing.T, iata, title, city, country string) airportDomain.Airport {
	t.Helper()
	c, err := common.NewCity(city)
	if err != nil {
		t.Fatalf("common.NewCity: err = %v", err)
	}
	ct, err := common.NewCountry(country)
	if err != nil {
		t.Fatalf("common.NewCountry: err = %v", err)
	}
	ic, err := airportDomain.NewIATACode(iata)
	if err != nil {
		t.Fatalf("airportDomain.NewIATACode: err = %v", err)
	}
	ti, err := airportDomain.NewTitle(title)
	if err != nil {
		t.Fatalf("airportDomain.NewTitle: err = %v", err)
	}
	a, err := airportDomain.NewAirport(ic, ti, c, ct)
	if err != nil {
		t.Fatalf("airportDomain.NewAirport: err = %v", err)
	}
	return a
}

func validFlight(t *testing.T, depGateID, arrGateID uuid.UUID) flightDomain.Flight {
	t.Helper()
	now := time.Now()
	f, err := flightDomain.NewFlight(
		uuid.New(),
		now,
		now.Add(time.Hour),
		nil,
		nil,
		"scheduled",
		nil,
		uuid.New(),
		uuid.New(),
		depGateID,
		arrGateID,
	)
	if err != nil {
		t.Fatalf("flightDomain.NewFlight: err = %v", err)
	}
	return f
}

func validRoute(t *testing.T, fid, depGateID, arrGateID uuid.UUID) flightDomain.FlightRoute {
	r, err := flightDomain.NewFlightRoute(fid, depGateID, arrGateID)
	if err != nil {
		t.Fatalf("flightDomain.NewFlightRoute: err = %v", err)
	}
	return r
}

func TestUserUsecase_GetUser(t *testing.T) {
	u := validUser(t)

	tests := []struct {
		name    string
		repo    userRepository.UserRepository
		email   string
		pass    string
		wantErr bool
	}{
		{
			name: "[P] valid credentials",
			repo: &userRepoMock{
				getUserFn: func(ctx context.Context, email string) (userDomain.User, error) {
					return u, nil
				},
			},
			email:   u.Email.String(),
			pass:    "Aa1!aaaa",
			wantErr: false,
		},
		{
			name: "[N] empty email",
			repo: &userRepoMock{
				getUserFn: func(ctx context.Context, email string) (userDomain.User, error) {
					return userDomain.User{}, errors.New("unexpected repo call")
				},
			},
			email:   "",
			pass:    "Aa1!aaaa",
			wantErr: true,
		},
		{
			name: "[N] wrong password",
			repo: &userRepoMock{
				getUserFn: func(ctx context.Context, email string) (userDomain.User, error) {
					return u, nil
				},
			},
			email:   u.Email.String(),
			pass:    "wrong-pass",
			wantErr: true,
		},
		{
			name: "[N] repo error",
			repo: &userRepoMock{
				getUserFn: func(ctx context.Context, email string) (userDomain.User, error) {
					return userDomain.User{}, errors.New("repo error")
				},
			},
			email:   u.Email.String(),
			pass:    "Aa1!aaaa",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			injector := do.New()
			do.Override(injector, func(i do.Injector) (userRepository.UserRepository, error) {
				return tt.repo, nil
			})
			do.Override(injector, func(i do.Injector) (publisherRepository.OutboxRepository, error) {
				return &outboxRepoMock{}, nil
			})
			do.Override(injector, func(i do.Injector) (flightRepository.FlightRepository, error) {
				return &flightRepoMock{}, nil
			})
			do.Override(injector, func(i do.Injector) (airportRepository.GateRepository, error) {
				return &gateRepoMock{}, nil
			})

			uc, err := NewUserUsecase(injector)
			if err != nil {
				t.Fatalf("NewUserUsecase: err = %v", err)
			}

			_, gotErr := uc.GetUser(tt.email, tt.pass)
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestUserUsecase_GetUserById(t *testing.T) {
	uid := uuid.New()
	u := validUser(t)

	tests := []struct {
		name    string
		repo    userRepository.UserRepository
		wantErr bool
	}{
		{
			name: "[P] found",
			repo: &userRepoMock{
				existFn: func(ctx context.Context, uid uuid.UUID) (userDomain.User, error) {
					return u, nil
				},
			},
			wantErr: false,
		},
		{
			name: "[N] not found",
			repo: &userRepoMock{
				existFn: func(ctx context.Context, uid uuid.UUID) (userDomain.User, error) {
					return userDomain.User{}, userRepository.ErrUserNotFound
				},
			},
			wantErr: true,
		},
		{
			name: "[N] repo error",
			repo: &userRepoMock{
				existFn: func(ctx context.Context, uid uuid.UUID) (userDomain.User, error) {
					return userDomain.User{}, errors.New("repo error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			injector := do.New()
			do.Override(injector, func(i do.Injector) (userRepository.UserRepository, error) {
				return tt.repo, nil
			})
			do.Override(injector, func(i do.Injector) (publisherRepository.OutboxRepository, error) {
				return &outboxRepoMock{}, nil
			})
			do.Override(injector, func(i do.Injector) (flightRepository.FlightRepository, error) {
				return &flightRepoMock{}, nil
			})
			do.Override(injector, func(i do.Injector) (airportRepository.GateRepository, error) {
				return &gateRepoMock{}, nil
			})

			uc, err := NewUserUsecase(injector)
			if err != nil {
				t.Fatalf("NewUserUsecase: err = %v", err)
			}

			_, gotErr := uc.GetUserById(uid)
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestUserUsecase_Subscribe(t *testing.T) {
	uid := uuid.New()
	fid := uuid.New()
	depGateID := uuid.New()
	arrGateID := uuid.New()
	u := validUser(t)
	depAirport := validAirport(t, "SVO", "Sheremetyevo", "Moscow", "RU")
	arrAirport := validAirport(t, "LED", "Pulkovo", "Saint Petersburg", "RU")
	f := validFlight(t, depGateID, arrGateID)
	r := validRoute(t, fid, depGateID, arrGateID)

	tests := []struct {
		name    string
		repo    userRepository.UserRepository
		wantErr bool
	}{
		{
			name: "[P] valid subscription",
			repo: &userRepoMock{
				subscribeFn: func(ctx context.Context, uid, fid uuid.UUID) error {
					return nil
				},
				existFn: func(ctx context.Context, uid uuid.UUID) (userDomain.User, error) {
					return u, nil
				},
			},
			wantErr: false,
		},
		{
			name: "[N] already subscribed",
			repo: &userRepoMock{
				subscribeFn: func(ctx context.Context, uid, fid uuid.UUID) error {
					return userRepository.ErrUserAlreadySubscribed
				},
			},
			wantErr: false,
		},
		{
			name: "[N] user not found",
			repo: &userRepoMock{
				subscribeFn: func(ctx context.Context, uid, fid uuid.UUID) error {
					return userRepository.ErrUserNotFound
				},
			},
			wantErr: true,
		},
		{
			name: "[N] flight not found",
			repo: &userRepoMock{
				subscribeFn: func(ctx context.Context, uid, fid uuid.UUID) error {
					return userRepository.ErrFlightNotFound
				},
			},
			wantErr: true,
		},
		{
			name: "[N] repo error",
			repo: &userRepoMock{
				subscribeFn: func(ctx context.Context, uid, fid uuid.UUID) error {
					return errors.New("repo error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saved := false
			injector := do.New()
			do.Override(injector, func(i do.Injector) (userRepository.UserRepository, error) {
				if repo, ok := tt.repo.(*userRepoMock); ok {
					if repo.existFn == nil {
						repo.existFn = func(ctx context.Context, uid uuid.UUID) (userDomain.User, error) {
							return u, nil
						}
					}
					if repo.listFn == nil {
						repo.listFn = func(ctx context.Context, uid uuid.UUID) ([]flightDomain.Flight, error) {
							return nil, nil
						}
					}
				}
				return tt.repo, nil
			})
			do.Override(injector, func(i do.Injector) (publisherRepository.OutboxRepository, error) {
				return &outboxRepoMock{
					saveFn: func(ctx context.Context, ob publisherDomain.Outbox) error {
						saved = true
						if ob.Topic != "subscription.created" {
							t.Fatalf("topic = %q, want = %q", ob.Topic, "subscription.created")
						}
						return nil
					},
				}, nil
			})
			do.Override(injector, func(i do.Injector) (flightRepository.FlightRepository, error) {
				return &flightRepoMock{
					existFn: func(ctx context.Context, fid uuid.UUID) (flightDomain.Flight, error) {
						return f, nil
					},
					routeFn: func(ctx context.Context, fid uuid.UUID) (flightDomain.FlightRoute, error) {
						return r, nil
					},
					listFn: func(ctx context.Context, fid uuid.UUID) ([]userDomain.User, error) {
						return []userDomain.User{u}, nil
					},
				}, nil
			})
			do.Override(injector, func(i do.Injector) (airportRepository.GateRepository, error) {
				return &gateRepoMock{
					airportFn: func(ctx context.Context, gid uuid.UUID) (airportDomain.Airport, error) {
						if gid == depGateID {
							return depAirport, nil
						}
						return arrAirport, nil
					},
				}, nil
			})
			t.Setenv("SUBSCRIPTION_CREATED_TOPIC", "subscription.created")

			uc, err := NewUserUsecase(injector)
			if err != nil {
				t.Fatalf("NewUserUsecase: err = %v", err)
			}

			gotErr := uc.Subscribe(uid, fid)
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
			}
			if tt.name == "[P] valid subscription" && !saved {
				t.Fatal("outbox was not saved")
			}
		})
	}
}

func TestUserUsecase_ListFlights(t *testing.T) {
	uid := uuid.New()

	tests := []struct {
		name    string
		repo    userRepository.UserRepository
		wantLen int
		wantErr bool
	}{
		{
			name: "[P] empty list",
			repo: &userRepoMock{
				listFn: func(ctx context.Context, uid uuid.UUID) ([]flightDomain.Flight, error) {
					return nil, nil
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "[P] non-empty list",
			repo: &userRepoMock{
				listFn: func(ctx context.Context, uid uuid.UUID) ([]flightDomain.Flight, error) {
					return []flightDomain.Flight{{}, {}}, nil
				},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "[N] repo error",
			repo: &userRepoMock{
				listFn: func(ctx context.Context, uid uuid.UUID) ([]flightDomain.Flight, error) {
					return nil, errors.New("repo error")
				},
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			injector := do.New()
			do.Override(injector, func(i do.Injector) (userRepository.UserRepository, error) {
				return tt.repo, nil
			})
			do.Override(injector, func(i do.Injector) (publisherRepository.OutboxRepository, error) {
				return &outboxRepoMock{}, nil
			})
			do.Override(injector, func(i do.Injector) (flightRepository.FlightRepository, error) {
				return &flightRepoMock{}, nil
			})
			do.Override(injector, func(i do.Injector) (airportRepository.GateRepository, error) {
				return &gateRepoMock{}, nil
			})

			uc, err := NewUserUsecase(injector)
			if err != nil {
				t.Fatalf("NewUserUsecase: err = %v", err)
			}

			got, gotErr := uc.ListFlights(uid)
			if gotErr != nil {
				if !tt.wantErr {
					t.Fatalf("gotErr = %v, wantErr = %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListFlights() succeeded unexpectedly")
			}
			if len(got) != tt.wantLen {
				t.Fatalf("got len = %d, want = %d", len(got), tt.wantLen)
			}
		})
	}
}

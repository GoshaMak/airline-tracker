package postgres

import (
	flightDomain "api/internal/flight/domain"
	flightModel "api/internal/flight/infra/postgres/model"
	"api/internal/pagination"
	"api/internal/user/domain"
	"api/internal/user/domain/repository"
	"api/internal/user/infra/postgres/model"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

type userRepository struct {
	conn userDB
}

type userDB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func NewUserRepository(i do.Injector) (repository.UserRepository, error) {
	return &userRepository{
		conn: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

func NewSubscriptionRepository(i do.Injector) (repository.SubscriptionRepository, error) {
	return &userRepository{
		conn: do.MustInvoke[*pgxpool.Pool](i),
	}, nil
}

func (r *userRepository) SaveUser(ctx context.Context, user domain.User) error {
	const op = "UserRepository.SaveUser"
	query := `
	insert into users(id, email, password_hash, role)
		values ($1, $2, $3, $4)
	`
	_, err := r.conn.Exec(ctx, query,
		user.Id,
		user.Email.String(),
		user.PasswordHash.String(),
		user.Role.String(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrUserAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *userRepository) GetUser(ctx context.Context, email string) (domain.User, error) {
	const op = "UserRepository.GetUser"
	query := `
	select * from users where email = $1
	`
	row, err := r.conn.Query(ctx, query, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}
	um, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[model.UserModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	ud, err := model.UserModelToDomain(um)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return ud, nil
}

func (r *userRepository) Exist(ctx context.Context, uid uuid.UUID) (domain.User, error) {
	const op = "UserRepository.Exist"
	query := `
	select * from users where id = $1
	`
	row, err := r.conn.Query(ctx, query, uid)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}
	um, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[model.UserModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	ud, err := model.UserModelToDomain(um)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return ud, nil
}

func (r *userRepository) Subscribe(
	ctx context.Context,
	uid,
	fid uuid.UUID,
) error {
	const op = "UserRepository.Subscribe"
	query := `
	select subscribe($1, $2)
	`
	_, err := r.conn.Exec(ctx, query, uid, fid)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "P0002" {
				if strings.HasPrefix(pgErr.Message, "user") {
					return repository.ErrUserNotFound
				} else if strings.HasPrefix(pgErr.Message, "flight") {
					return repository.ErrFlightNotFound
				}
			} else if pgErr.Code == "23505" &&
				pgErr.ConstraintName == "unique_flight_subscription_per_user" {
				return repository.ErrUserAlreadySubscribed
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *userRepository) Create(
	ctx context.Context,
	uid,
	fid uuid.UUID,
	notifyBeforeMinutes int64,
) (bool, error) {
	const op = "UserRepository.CreateSubscription"
	if err := r.Subscribe(ctx, uid, fid); err != nil &&
		!errors.Is(err, repository.ErrUserAlreadySubscribed) {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	query := `
	insert into subscription_notification_timers(subscription_id, notify_before_minutes)
	select id, $3
	from subscriptions
	where user_id = $1 and flight_id = $2
	on conflict (subscription_id, notify_before_minutes) do nothing
	returning notify_before_minutes
	`
	rows, err := r.conn.Query(ctx, query, uid, fid, notifyBeforeMinutes)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	values, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return len(values) > 0, nil
}

func (r *userRepository) AddTimers(
	ctx context.Context,
	uid,
	fid uuid.UUID,
	notifyBeforeMinutes []int64,
) ([]int64, error) {
	const op = "UserRepository.AddTimers"
	findQuery := `
	select id
	from subscriptions
	where user_id = $1 and flight_id = $2
	`
	rows, err := r.conn.Query(ctx, findQuery, uid, fid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	subscriptionID, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	insertQuery := `
	insert into subscription_notification_timers(subscription_id, notify_before_minutes)
	select $1, unnest($2::bigint[])
	on conflict (subscription_id, notify_before_minutes) do nothing
	returning notify_before_minutes
	`
	rows, err = r.conn.Query(ctx, insertQuery, subscriptionID, notifyBeforeMinutes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	added, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return added, nil
}

func (r *userRepository) Delete(ctx context.Context, uid, fid uuid.UUID) error {
	const op = "UserRepository.DeleteSubscription"
	query := `
	delete from subscriptions
	where user_id = $1 and flight_id = $2
	`
	if _, err := r.conn.Exec(ctx, query, uid, fid); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *userRepository) ListFlights(
	ctx context.Context,
	uid uuid.UUID,
) ([]flightDomain.Flight, error) {
	const op = "UserRepository.ListFlights"
	query := `
	select * from scan_user_flights_info($1)
	`
	rows, err := r.conn.Query(ctx, query, uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	fsMs, err := pgx.CollectRows(rows, pgx.RowToStructByName[flightModel.FlightModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	fsDs := make([]flightDomain.Flight, len(fsMs))
	for i, fm := range fsMs {
		fd, err := flightModel.FlightModelToDomain(fm)
		if err != nil {
			return nil, err
		}
		fsDs[i] = fd
	}
	return fsDs, nil
}

func (r *userRepository) ListFlightsPage(
	ctx context.Context,
	uid uuid.UUID,
	params pagination.Params,
) ([]flightDomain.Flight, error) {
	const op = "UserRepository.ListFlightsPage"
	query := `
	select *
	from scan_user_flights_info($1)
	where ($2::uuid is null or id > $2)
	order by id
	limit $3
	`
	rows, err := r.conn.Query(ctx, query, uid, params.CursorValue(), params.FetchLimit())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[flightModel.FlightModel])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	flights := make([]flightDomain.Flight, len(models))
	for i, modelItem := range models {
		flight, err := flightModel.FlightModelToDomain(modelItem)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		flights[i] = flight
	}
	return flights, nil
}

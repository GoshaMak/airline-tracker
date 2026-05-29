package redis

import (
	"api/internal/flight/domain"
	"api/internal/flight/infra/redis/model"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	rds "github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
)

type RedisDB struct {
	cln *rds.Client
}

func NewRedisDB(i do.Injector) (*RedisDB, error) {
	return &RedisDB{
		cln: do.MustInvoke[*rds.Client](i),
	}, nil
}

func formFlightKey(id string) string {
	return "flight:" + id
}

func flightIndexKey() string {
	return "flight:ids"
}

func (r *RedisDB) SaveFlight(
	ctx context.Context,
	flight domain.Flight,
) error {
	const op = "RedisDB.SaveFlight"
	fm, err := model.FlightDomainToModel(flight)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	fKey := formFlightKey(flight.Id.String())
	fiKey := flightIndexKey()
	tx := r.cln.TxPipeline()
	tx.HSet(ctx, fKey, fm)
	tx.SAdd(ctx, fiKey, fKey)
	if _, err := tx.Exec(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *RedisDB) SaveFlights(
	ctx context.Context,
	flights []domain.Flight,
) error {
	const op = "RedisDB.SaveFlights"
	tx := r.cln.TxPipeline()
	for _, f := range flights {
		fKey := formFlightKey(f.Id.String())
		fm, err := model.FlightDomainToModel(f)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		fiKey := flightIndexKey()
		tx.HSet(ctx, fKey, fm)
		tx.SAdd(ctx, fiKey, fKey)
	}
	if _, err := tx.Exec(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *RedisDB) GetFlightById(
	ctx context.Context,
	fid uuid.UUID,
) (domain.Flight, error) {
	const op = "RedisDB.GetFlightById"
	var fm model.FlightModel
	fKey := formFlightKey(fid.String())
	cmd := r.cln.HGetAll(ctx, fKey)
	result, err := cmd.Result()
	if err != nil {
		if err == rds.Nil {
			return domain.Flight{}, ErrFlightNotFound
		}
		return domain.Flight{}, fmt.Errorf("%s: %w", op, err)
	}
	if len(result) == 0 {
		return domain.Flight{}, ErrFlightNotFound
	}
	if err := cmd.Scan(&fm); err != nil {
		return domain.Flight{}, fmt.Errorf("%s: %w", op, err)
	}
	fm.Id = fid
	f, err := model.FlightModelToDomain(fm)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("%s: %w", op, err)
	}
	return f, nil
}

func (r *RedisDB) DeleteFlightById(
	ctx context.Context,
	fid uuid.UUID,
) error {
	const op = "RedisDB.DeleteFlightById"
	fKey := formFlightKey(fid.String())
	fiKey := flightIndexKey()
	tx := r.cln.TxPipeline()
	tx.SRem(ctx, fiKey, fKey)
	tx.Unlink(ctx, fKey)
	if _, err := tx.Exec(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// flights must be already cached
func (r *RedisDB) UpdateFlights(ctx context.Context, flights []domain.Flight) error {
	const op = "RedisDB.UpdateFlights"
	pipe := r.cln.Pipeline()
	for _, f := range flights {
		fKey := formFlightKey(f.Id.String())
		fm, err := model.FlightDomainToModel(f)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		pipe.HSet(ctx, fKey, fm)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *RedisDB) GetFlights(ctx context.Context) ([]domain.Flight, error) {
	const op = "RedisDB.GetFlights"
	pipe := r.cln.Pipeline()
	fiKey := flightIndexKey()
	iter := r.cln.SScan(ctx, fiKey, 0, "", 0).Iterator()
	for iter.Next(ctx) {
		fKey := iter.Val()
		pipe.HGetAll(ctx, fKey)
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	fs := make([]domain.Flight, 0, len(cmds))
	for _, rawCmd := range cmds {
		cmd, ok := rawCmd.(*rds.MapStringStringCmd)
		if !ok {
			return nil, fmt.Errorf("%s: unexpected command type %T", op, rawCmd)
		}

		r, err := cmd.Result()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if len(r) == 0 {
			return nil, fmt.Errorf("%s: empty flight hash", op)
		}

		var fm model.FlightModel
		if err := cmd.Scan(&fm); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		key := strings.TrimPrefix(fmt.Sprint(cmd.Args()[1]), "flight:")
		id, err := uuid.Parse(key)
		if err != nil {
			return nil, fmt.Errorf("%s: parse flight id from cache key %q: %w", op, key, err)
		}
		fm.Id = id

		f, err := model.FlightModelToDomain(fm)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		fs = append(fs, f)
	}
	return fs, nil
}

func (r *RedisDB) UpdateFlight(
	ctx context.Context,
	ufi domain.UpdateFlightInfo,
) error {
	const op = "RedisDB.UpdateFlight"
	ufim, err := model.UpdateFlightInfoDomainToModel(ufi)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	fKey := formFlightKey(ufi.FlightId.String())
	if err := r.cln.HSet(ctx, fKey, ufim).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *RedisDB) FlushFlights(ctx context.Context) error {
	const op = "RedisDB.FlushFlights"
	const batchSize = 1000
	iter := r.cln.Scan(ctx, 0, "flight:*", 0).Iterator()
	pipe := r.cln.Pipeline()

	batch := 0
	exec := func() error {
		if batch == 0 {
			return nil
		}
		if _, err := pipe.Exec(ctx); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		batch = 0
		pipe = r.cln.Pipeline()
		return nil
	}

	fiKey := flightIndexKey()
	pipe.Unlink(ctx, fiKey)
	batch++
	for iter.Next(ctx) {
		fKey := iter.Val()
		pipe.Unlink(ctx, fKey)

		batch++
		if batch == batchSize {
			if err := exec(); err != nil {
				return err
			}
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("%s: iterator error %w", op, err)
	}
	if err := exec(); err != nil {
		return err
	}
	return nil
}

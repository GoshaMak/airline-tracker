package redis

import (
	"api/internal/flight/domain"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	rds "github.com/redis/go-redis/v9"
)

// flightMother supplies a valid object; no Redis server is needed.
func flightMother() domain.Flight {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	return domain.Flight{Id: uuid.New(), AircraftId: uuid.New(),
		ScheduledDeparture: now, ScheduledArrival: now.Add(time.Hour),
		Status: domain.Scheduled, DepartureAirportId: uuid.New(), ArrivalAirportId: uuid.New(),
		DepartureGateId: uuid.New(), ArrivalGateId: uuid.New()}
}
func flightHash(f domain.Flight) map[string]string {
	return map[string]string{
		"aircraft_id": f.AircraftId.String(), "scheduled_departure": f.ScheduledDeparture.Format(time.RFC3339Nano),
		"scheduled_arrival": f.ScheduledArrival.Format(time.RFC3339Nano), "status": "scheduled",
		"departure_airport_id": f.DepartureAirportId.String(), "arrival_airport_id": f.ArrivalAirportId.String(),
		"departure_gate_id": f.DepartureGateId.String(), "arrival_gate_id": f.ArrivalGateId.String(),
	}
}

type fakeRedis struct {
	redisClient
	hashes  map[string]map[string]string
	keys    []string
	pipeErr error
	hsetErr error
	scanErr error
	calls   []string
}

func newFakeRedis() *fakeRedis                 { return &fakeRedis{hashes: make(map[string]map[string]string)} }
func (f *fakeRedis) Pipeline() rds.Pipeliner   { return &fakePipe{f: f} }
func (f *fakeRedis) TxPipeline() rds.Pipeliner { return &fakePipe{f: f} }
func (f *fakeRedis) HGetAll(ctx context.Context, key string) *rds.MapStringStringCmd {
	f.calls = append(f.calls, "HGETALL "+key)
	c := rds.NewMapStringStringCmd(ctx, "hgetall", key)
	c.SetVal(f.hashes[key])
	return c
}
func (f *fakeRedis) HSet(ctx context.Context, key string, _ ...interface{}) *rds.IntCmd {
	f.calls = append(f.calls, "HSET "+key)
	c := rds.NewIntCmd(ctx, "hset", key)
	if f.hsetErr != nil {
		c.SetErr(f.hsetErr)
	}
	return c
}
func (f *fakeRedis) SScan(ctx context.Context, _ string, _ uint64, _ string, _ int64) *rds.ScanCmd {
	c := rds.NewScanCmd(ctx, nil, "sscan")
	c.SetVal(f.keys, 0)
	if f.scanErr != nil {
		c.SetErr(f.scanErr)
	}
	return c
}
func (f *fakeRedis) Scan(ctx context.Context, _ uint64, _ string, _ int64) *rds.ScanCmd {
	c := rds.NewScanCmd(ctx, nil, "scan")
	c.SetVal(f.keys, 0)
	if f.scanErr != nil {
		c.SetErr(f.scanErr)
	}
	return c
}

type fakePipe struct {
	rds.Pipeliner
	f    *fakeRedis
	cmds []rds.Cmder
}

func (p *fakePipe) HSet(ctx context.Context, key string, _ ...interface{}) *rds.IntCmd {
	p.f.calls = append(p.f.calls, "HSET "+key)
	c := rds.NewIntCmd(ctx, "hset", key)
	p.cmds = append(p.cmds, c)
	return c
}
func (p *fakePipe) SAdd(ctx context.Context, key string, _ ...interface{}) *rds.IntCmd {
	p.f.calls = append(p.f.calls, "SADD "+key)
	c := rds.NewIntCmd(ctx, "sadd", key)
	p.cmds = append(p.cmds, c)
	return c
}
func (p *fakePipe) SRem(ctx context.Context, key string, _ ...interface{}) *rds.IntCmd {
	p.f.calls = append(p.f.calls, "SREM "+key)
	c := rds.NewIntCmd(ctx, "srem", key)
	p.cmds = append(p.cmds, c)
	return c
}
func (p *fakePipe) Unlink(ctx context.Context, key ...string) *rds.IntCmd {
	p.f.calls = append(p.f.calls, "UNLINK "+key[0])
	c := rds.NewIntCmd(ctx, "unlink", key[0])
	p.cmds = append(p.cmds, c)
	return c
}
func (p *fakePipe) HGetAll(ctx context.Context, key string) *rds.MapStringStringCmd {
	p.f.calls = append(p.f.calls, "HGETALL "+key)
	c := rds.NewMapStringStringCmd(ctx, "hgetall", key)
	c.SetVal(p.f.hashes[key])
	p.cmds = append(p.cmds, c)
	return c
}
func (p *fakePipe) Exec(context.Context) ([]rds.Cmder, error) { return p.cmds, p.f.pipeErr }

func TestRedisDB_SaveFlight(t *testing.T) {
	for _, tt := range []struct {
		name    string
		failure error
	}{{"positive", nil}, {"negative pipeline error", errors.New("redis unavailable")}} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			f := flightMother()
			fake := newFakeRedis()
			fake.pipeErr = tt.failure
			repo := &RedisDB{cln: fake}
			// Act
			err := repo.SaveFlight(context.Background(), f)
			// Assert
			if (tt.failure == nil && err != nil) || (tt.failure != nil && !errors.Is(err, tt.failure)) {
				t.Fatal(err)
			}
			if len(fake.calls) != 2 || fake.calls[0] != "HSET flight:"+f.Id.String() || fake.calls[1] != "SADD flight:ids" {
				t.Fatalf("calls=%v", fake.calls)
			}
		})
	}
}
func TestRedisDB_SaveFlights(t *testing.T) {
	for _, tt := range []struct {
		name    string
		failure error
	}{{"positive", nil}, {"negative pipeline error", errors.New("redis unavailable")}} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			f := flightMother()
			fake := newFakeRedis()
			fake.pipeErr = tt.failure
			repo := &RedisDB{cln: fake}
			// Act
			err := repo.SaveFlights(context.Background(), []domain.Flight{f})
			// Assert
			if (tt.failure == nil && err != nil) || (tt.failure != nil && !errors.Is(err, tt.failure)) {
				t.Fatal(err)
			}
			if len(fake.calls) != 2 || fake.calls[0] != "HSET flight:"+f.Id.String() {
				t.Fatalf("calls=%v", fake.calls)
			}
		})
	}
}
func TestRedisDB_GetFlightById(t *testing.T) {
	t.Run("positive: decodes cached flight", func(t *testing.T) {
		// Arrange
		f := flightMother()
		fake := newFakeRedis()
		fake.hashes["flight:"+f.Id.String()] = flightHash(f)
		repo := &RedisDB{cln: fake}
		// Act
		got, err := repo.GetFlightById(context.Background(), f.Id)
		// Assert
		if err != nil || got.Id != f.Id || got.Status != domain.Scheduled {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})
	t.Run("negative: missing hash", func(t *testing.T) {
		// Arrange
		repo := &RedisDB{cln: newFakeRedis()}
		// Act
		_, err := repo.GetFlightById(context.Background(), uuid.New())
		// Assert
		if !errors.Is(err, ErrFlightNotFound) {
			t.Fatal(err)
		}
	})
}
func TestRedisDB_DeleteFlightById(t *testing.T) {
	for _, tt := range []struct {
		name    string
		failure error
	}{{"positive", nil}, {"negative pipeline error", errors.New("redis unavailable")}} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			f := flightMother()
			fake := newFakeRedis()
			fake.pipeErr = tt.failure
			repo := &RedisDB{cln: fake}
			// Act
			err := repo.DeleteFlightById(context.Background(), f.Id)
			// Assert
			if (tt.failure == nil && err != nil) || (tt.failure != nil && !errors.Is(err, tt.failure)) {
				t.Fatal(err)
			}
			if len(fake.calls) != 2 || fake.calls[0] != "SREM flight:ids" || fake.calls[1] != "UNLINK flight:"+f.Id.String() {
				t.Fatalf("calls=%v", fake.calls)
			}
		})
	}
}
func TestRedisDB_UpdateFlights(t *testing.T) {
	for _, tt := range []struct {
		name    string
		failure error
	}{{"positive", nil}, {"negative pipeline error", errors.New("redis unavailable")}} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			f := flightMother()
			fake := newFakeRedis()
			fake.pipeErr = tt.failure
			repo := &RedisDB{cln: fake}
			// Act
			err := repo.UpdateFlights(context.Background(), []domain.Flight{f})
			// Assert
			if (tt.failure == nil && err != nil) || (tt.failure != nil && !errors.Is(err, tt.failure)) {
				t.Fatal(err)
			}
			if len(fake.calls) != 1 || fake.calls[0] != "HSET flight:"+f.Id.String() {
				t.Fatalf("calls=%v", fake.calls)
			}
		})
	}
}
func TestRedisDB_GetFlights(t *testing.T) {
	t.Run("positive: reads indexed flight", func(t *testing.T) {
		// Arrange
		f := flightMother()
		fake := newFakeRedis()
		key := "flight:" + f.Id.String()
		fake.keys = []string{key}
		fake.hashes[key] = flightHash(f)
		repo := &RedisDB{cln: fake}
		// Act
		got, err := repo.GetFlights(context.Background())
		// Assert
		if err != nil || len(got) != 1 || got[0].Id != f.Id {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})
	t.Run("negative: index scan fails", func(t *testing.T) {
		// Arrange
		failure := errors.New("redis unavailable")
		fake := newFakeRedis()
		fake.scanErr = failure
		repo := &RedisDB{cln: fake}
		// Act
		_, err := repo.GetFlights(context.Background())
		// Assert
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}
func TestRedisDB_UpdateFlight(t *testing.T) {
	for _, tt := range []struct {
		name    string
		failure error
	}{{"positive", nil}, {"negative HSET error", errors.New("redis unavailable")}} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			f := flightMother()
			fake := newFakeRedis()
			fake.hsetErr = tt.failure
			repo := &RedisDB{cln: fake}
			status := "boarding"
			// Act
			err := repo.UpdateFlight(context.Background(), domain.UpdateFlightInfo{FlightId: f.Id, Status: &status})
			// Assert
			if (tt.failure == nil && err != nil) || (tt.failure != nil && !errors.Is(err, tt.failure)) {
				t.Fatal(err)
			}
			if len(fake.calls) != 1 || fake.calls[0] != "HSET flight:"+f.Id.String() {
				t.Fatalf("calls=%v", fake.calls)
			}
		})
	}
}
func TestRedisDB_FlushFlights(t *testing.T) {
	t.Run("positive: removes index and cached hashes", func(t *testing.T) {
		// Arrange
		f := flightMother()
		fake := newFakeRedis()
		fake.keys = []string{"flight:" + f.Id.String()}
		repo := &RedisDB{cln: fake}
		// Act
		err := repo.FlushFlights(context.Background())
		// Assert
		if err != nil || len(fake.calls) != 2 || fake.calls[0] != "UNLINK flight:ids" || fake.calls[1] != "UNLINK flight:"+f.Id.String() {
			t.Fatalf("err=%v, calls=%v", err, fake.calls)
		}
	})
	t.Run("negative: pipeline error", func(t *testing.T) {
		// Arrange
		failure := errors.New("redis unavailable")
		fake := newFakeRedis()
		fake.pipeErr = failure
		repo := &RedisDB{cln: fake}
		// Act
		err := repo.FlushFlights(context.Background())
		// Assert
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

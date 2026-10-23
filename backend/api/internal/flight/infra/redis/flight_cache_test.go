package redis

import (
	"api/internal/flight/domain"
	"api/internal/flight/infra/redis/model"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
)

func flightMother() domain.Flight {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	return domain.Flight{
		Id:                 uuid.New(),
		AircraftId:         uuid.New(),
		ScheduledDeparture: now,
		ScheduledArrival:   now.Add(time.Hour),
		Status:             domain.Scheduled,
		DepartureAirportId: uuid.New(),
		ArrivalAirportId:   uuid.New(),
		DepartureGateId:    uuid.New(),
		ArrivalGateId:      uuid.New(),
	}
}

func flightHash(f domain.Flight) map[string]string {
	hash := map[string]string{
		"aircraft_id":          f.AircraftId.String(),
		"scheduled_departure":  f.ScheduledDeparture.Format(time.RFC3339Nano),
		"scheduled_arrival":    f.ScheduledArrival.Format(time.RFC3339Nano),
		"status":               f.Status.String(),
		"departure_airport_id": f.DepartureAirportId.String(),
		"arrival_airport_id":   f.ArrivalAirportId.String(),
		"departure_gate_id":    f.DepartureGateId.String(),
		"arrival_gate_id":      f.ArrivalGateId.String(),
	}
	if f.ActualDeparture != nil {
		hash["actual_departure"] = f.ActualDeparture.Format(time.RFC3339Nano)
	}
	if f.ActualArrival != nil {
		hash["actual_arrival"] = f.ActualArrival.Format(time.RFC3339Nano)
	}
	if f.Plan != nil {
		hash["plan"] = f.Plan.String()
	}
	return hash
}

func newRedisFixture(t *testing.T) (*RedisClient, redismock.ClientMock) {
	t.Helper()
	client, mock := redismock.NewClientMock()
	t.Cleanup(func() { _ = client.Close() })
	return &RedisClient{cln: client}, mock
}

func assertRedisExpectations(t *testing.T, mock redismock.ClientMock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("redis expectations: %v", err)
	}
}

func expectFlightHSet(mock redismock.ClientMock, f domain.Flight) *redismock.ExpectedInt {
	var plan *string
	if f.Plan != nil {
		plan = new(f.Plan.String())
	}
	return mock.ExpectHSet("flight:"+f.Id.String(), model.FlightModel{
		Id:                 f.Id,
		AircraftId:         f.AircraftId.String(),
		ScheduledDeparture: f.ScheduledDeparture,
		ScheduledArrival:   f.ScheduledArrival,
		ActualDeparture:    f.ActualDeparture,
		ActualArrival:      f.ActualArrival,
		Status:             f.Status.String(),
		Plan:               plan,
		DepartureAirportId: f.DepartureAirportId.String(),
		ArrivalAirportId:   f.ArrivalAirportId.String(),
		DepartureGateId:    f.DepartureGateId.String(),
		ArrivalGateId:      f.ArrivalGateId.String(),
	})
}

func TestRedisDB_SaveFlight(t *testing.T) {
	for _, tt := range []struct {
		name    string
		failure error
	}{
		{
			name: "positive",
		},
		{
			name:    "negative pipeline error",
			failure: errors.New("redis unavailable"),
		},
	} {
		allure.Test(t, tt.name, func(ctx *allure.Context) {
			t := ctx.T()
			repo, mock := newRedisFixture(t)
			defer assertRedisExpectations(t, mock)
			f := flightMother()
			f.ActualDeparture = new(f.ScheduledDeparture.Add(time.Minute))
			f.ActualArrival = new(f.ScheduledArrival.Add(time.Minute))
			f.Plan = new(domain.FlightPlan("DCT SVO LED"))
			mock.ExpectTxPipeline()
			expectFlightHSet(mock, f).SetVal(1)
			mock.ExpectSAdd("flight:ids", "flight:"+f.Id.String()).SetVal(1)
			exec := mock.ExpectTxPipelineExec()
			if tt.failure != nil {
				exec.SetErr(tt.failure)
			}
			err := repo.SaveFlight(context.Background(), f)
			if !errors.Is(err, tt.failure) {
				t.Fatalf("err=%v, want=%v", err, tt.failure)
			}
		})
	}
}

func TestRedisDB_SaveFlights(t *testing.T) {
	for _, tt := range []struct {
		name    string
		failure error
	}{
		{
			name: "positive",
		},
		{
			name:    "negative pipeline error",
			failure: errors.New("redis unavailable"),
		},
	} {
		allure.Test(t, tt.name, func(ctx *allure.Context) {
			t := ctx.T()
			repo, mock := newRedisFixture(t)
			defer assertRedisExpectations(t, mock)
			f, second := flightMother(), flightMother()
			second.Status = domain.Boarding
			second.Plan = new(domain.FlightPlan("DCT SVO LED"))
			flights := []domain.Flight{f, second}
			mock.ExpectTxPipeline()
			for _, flight := range flights {
				expectFlightHSet(mock, flight).SetVal(1)
				mock.ExpectSAdd("flight:ids", "flight:"+flight.Id.String()).SetVal(1)
			}
			exec := mock.ExpectTxPipelineExec()
			if tt.failure != nil {
				exec.SetErr(tt.failure)
			}
			err := repo.SaveFlights(context.Background(), flights)
			if !errors.Is(err, tt.failure) {
				t.Fatalf("err=%v, want=%v", err, tt.failure)
			}
		})
	}
}

func TestRedisDB_GetFlightById(t *testing.T) {
	allure.Test(t, "positive: decodes cached flight", func(ctx *allure.Context) {
		t := ctx.T()
		f := flightMother()
		f.ActualDeparture = new(f.ScheduledDeparture.Add(time.Minute))
		f.ActualArrival = new(f.ScheduledArrival.Add(time.Minute))
		f.Plan = new(domain.FlightPlan("DCT SVO LED"))
		repo, mock := newRedisFixture(t)
		defer assertRedisExpectations(t, mock)
		mock.ExpectHGetAll("flight:" + f.Id.String()).SetVal(flightHash(f))
		got, err := repo.GetFlightById(context.Background(), f.Id)
		if err != nil || !reflect.DeepEqual(got, f) {
			t.Fatalf("got=%+v, want=%+v, err=%v", got, f, err)
		}
	})

	allure.Test(t, "negative: missing hash", func(ctx *allure.Context) {
		t := ctx.T()
		fid := uuid.New()
		repo, mock := newRedisFixture(t)
		defer assertRedisExpectations(t, mock)
		mock.ExpectHGetAll("flight:" + fid.String()).SetVal(map[string]string{})
		_, err := repo.GetFlightById(context.Background(), fid)
		if !errors.Is(err, ErrFlightNotFound) {
			t.Fatal(err)
		}
	})
}

func TestRedisDB_DeleteFlightById(t *testing.T) {
	for _, tt := range []struct {
		name    string
		failure error
	}{
		{
			name: "positive",
		},
		{
			name:    "negative pipeline error",
			failure: errors.New("redis unavailable"),
		},
	} {
		allure.Test(t, tt.name, func(ctx *allure.Context) {
			t := ctx.T()
			repo, mock := newRedisFixture(t)
			defer assertRedisExpectations(t, mock)
			fid := uuid.New()
			key := "flight:" + fid.String()
			mock.ExpectTxPipeline()
			mock.ExpectSRem("flight:ids", key).SetVal(1)
			mock.ExpectUnlink(key).SetVal(1)
			exec := mock.ExpectTxPipelineExec()
			if tt.failure != nil {
				exec.SetErr(tt.failure)
			}
			err := repo.DeleteFlightById(context.Background(), fid)
			if !errors.Is(err, tt.failure) {
				t.Fatalf("err=%v, want=%v", err, tt.failure)
			}
		})
	}
}

func TestRedisDB_UpdateFlights(t *testing.T) {
	for _, tt := range []struct {
		name    string
		failure error
	}{
		{
			name: "positive",
		},
		{
			name:    "negative pipeline error",
			failure: errors.New("redis unavailable"),
		},
	} {
		allure.Test(t, tt.name, func(ctx *allure.Context) {
			t := ctx.T()
			repo, mock := newRedisFixture(t)
			defer assertRedisExpectations(t, mock)
			f, second := flightMother(), flightMother()
			second.Status = domain.Boarding
			second.Plan = new(domain.FlightPlan("DCT SVO LED"))
			flights := []domain.Flight{f, second}
			expectFlightHSet(mock, f).SetVal(1)
			last := expectFlightHSet(mock, second)
			if tt.failure != nil {
				last.SetErr(tt.failure)
			} else {
				last.SetVal(1)
			}
			err := repo.UpdateFlights(context.Background(), flights)
			if !errors.Is(err, tt.failure) {
				t.Fatalf("err=%v, want=%v", err, tt.failure)
			}
		})
	}
}

func TestRedisDB_GetFlights(t *testing.T) {
	allure.Test(t, "positive: reads indexed flight", func(ctx *allure.Context) {
		t := ctx.T()
		f := flightMother()
		key := "flight:" + f.Id.String()
		repo, mock := newRedisFixture(t)
		defer assertRedisExpectations(t, mock)
		mock.ExpectSScan("flight:ids", 0, "", 0).SetVal([]string{key}, 0)
		mock.ExpectHGetAll(key).SetVal(flightHash(f))
		got, err := repo.GetFlights(context.Background())
		if err != nil || !reflect.DeepEqual(got, []domain.Flight{f}) {
			t.Fatalf("got=%+v, err=%v", got, err)
		}
	})

	allure.Test(t, "negative: index scan fails", func(ctx *allure.Context) {
		t := ctx.T()
		failure := errors.New("redis unavailable")
		repo, mock := newRedisFixture(t)
		defer assertRedisExpectations(t, mock)
		mock.ExpectSScan("flight:ids", 0, "", 0).SetErr(failure)
		_, err := repo.GetFlights(context.Background())
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

func TestRedisDB_UpdateFlight(t *testing.T) {
	for _, tt := range []struct {
		name    string
		failure error
	}{
		{
			name: "positive",
		},
		{
			name:    "negative HSET error",
			failure: errors.New("redis unavailable"),
		},
	} {
		allure.Test(t, tt.name, func(ctx *allure.Context) {
			t := ctx.T()
			repo, mock := newRedisFixture(t)
			defer assertRedisExpectations(t, mock)
			fid := uuid.New()
			status := "boarding"
			update := domain.UpdateFlightInfo{FlightId: fid, Status: &status}
			expected := mock.ExpectHSet("flight:"+fid.String(), model.UpdateFlightInfoModel{
				FlightId: fid,
				Status:   &status,
			})
			if tt.failure != nil {
				expected.SetErr(tt.failure)
			} else {
				expected.SetVal(1)
			}
			err := repo.UpdateFlight(context.Background(), update)
			if !errors.Is(err, tt.failure) {
				t.Fatalf("err=%v, want=%v", err, tt.failure)
			}
		})
	}
}

func TestRedisDB_FlushFlights(t *testing.T) {
	allure.Test(t, "positive: removes index and cached hashes", func(ctx *allure.Context) {
		t := ctx.T()
		key := "flight:" + uuid.NewString()
		repo, mock := newRedisFixture(t)
		defer assertRedisExpectations(t, mock)
		mock.ExpectScan(0, "flight:*", 0).SetVal([]string{key}, 0)
		mock.ExpectUnlink("flight:ids").SetVal(1)
		mock.ExpectUnlink(key).SetVal(1)
		err := repo.FlushFlights(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	})

	allure.Test(t, "negative: pipeline error", func(ctx *allure.Context) {
		t := ctx.T()
		failure := errors.New("redis unavailable")
		repo, mock := newRedisFixture(t)
		defer assertRedisExpectations(t, mock)
		mock.ExpectScan(0, "flight:*", 0).SetVal([]string{}, 0)
		mock.ExpectUnlink("flight:ids").SetErr(failure)
		err := repo.FlushFlights(context.Background())
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

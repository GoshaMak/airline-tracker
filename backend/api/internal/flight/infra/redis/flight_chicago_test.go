package redis

import (
	"api/internal/flight/domain"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	allure "github.com/allure-framework/allure-go/commons/gotest"
	rds "github.com/redis/go-redis/v9"
)

func newRedisChicagoFixture(t *testing.T) (*RedisClient, *miniredis.Miniredis) {
	t.Helper()
	redis := miniredis.RunT(t)
	client := rds.NewClient(&rds.Options{Addr: redis.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return &RedisClient{cln: client}, redis
}

func TestRedisDB_UpdateFlight_Chicago(t *testing.T) {
	allure.Test(t, "positive: persists changes and preserves another flight", func(ctx *allure.Context) {
		t := ctx.T()

		repo, server := newRedisChicagoFixture(t)
		flight, other := flightMother(), flightMother()
		if err := repo.SaveFlights(context.Background(), []domain.Flight{flight, other}); err != nil {
			t.Fatalf("seed flights: %v", err)
		}
		status, plan := "boarding", "DCT SVO LED"
		departure := flight.ScheduledDeparture.Add(5 * time.Minute)
		update := domain.UpdateFlightInfo{
			FlightId:        flight.Id,
			ActualDeparture: &departure,
			Status:          &status,
			Plan:            &plan,
		}
		want := flight
		want.ActualDeparture = &departure
		want.Status = domain.Boarding
		want.Plan = new(domain.FlightPlan(plan))

		err := repo.UpdateFlight(context.Background(), update)

		if err != nil {
			t.Fatalf("update flight: %v", err)
		}
		got, err := repo.GetFlightById(context.Background(), flight.Id)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("updated flight=%+v, want=%+v, err=%v", got, want, err)
		}
		unchanged, err := repo.GetFlightById(context.Background(), other.Id)
		if err != nil || !reflect.DeepEqual(unchanged, other) {
			t.Fatalf("other flight=%+v, want=%+v, err=%v", unchanged, other, err)
		}
		key := "flight:" + flight.Id.String()
		if server.HGet(key, "status") != status || server.HGet(key, "plan") != plan {
			t.Fatal("updated status and plan were not persisted")
		}
	})
}

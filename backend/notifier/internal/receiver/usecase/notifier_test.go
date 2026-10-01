package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"notifier/internal/receiver/command"
	"notifier/internal/receiver/domain"
	"shared/common"
	"testing"
	"time"

	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/google/uuid"
)

// subscriptionMother supplies a complete, valid command for every scenario.
func subscriptionMother(t *testing.T) command.SubscriptionCreatedCommand {
	t.Helper()
	email, err := common.NewEmail("traveler@example.com")
	if err != nil {
		t.Fatal(err)
	}
	city, err := common.NewCity("Moscow")
	if err != nil {
		t.Fatal(err)
	}
	country, err := common.NewCountry("RU")
	if err != nil {
		t.Fatal(err)
	}
	return command.SubscriptionCreatedCommand{
		Email: email, ScheduledDeparture: time.Now().UTC().Add(2 * time.Hour),
		FlightStatus: "scheduled", DepartureAirportIATACode: "SVO",
		DepartureAirportTitle: "Sheremetyevo", DepartureAirportCity: city,
		DepartureAirportCountry: country, ArrivalAirportIATACode: "LED",
		ArrivalAirportTitle: "Pulkovo", ArrivalAirportCity: city,
		ArrivalAirportCountry: country,
	}
}

// updateBuilder makes one valid update and lets a test change only its relevant field.
type updateBuilder struct{ cmd command.FlightUpdatedCommand }

func newUpdateBuilder(t *testing.T) updateBuilder {
	t.Helper()
	email, err := common.NewEmail("traveler@example.com")
	if err != nil {
		t.Fatal(err)
	}
	status := "boarding"
	return updateBuilder{command.FlightUpdatedCommand{
		FlightId: uuid.New(), Users: []common.Email{email},
		DepartureAirportTitle: "Sheremetyevo", ArrivalAirportTitle: "Pulkovo",
		Status: &status,
	}}
}
func (b updateBuilder) withoutUsers() updateBuilder         { b.cmd.Users = nil; return b }
func (b updateBuilder) build() command.FlightUpdatedCommand { return b.cmd }

// memoryNotifications is a classic in-memory implementation of the repository.
type memoryNotifications struct {
	items []domain.Notification
	err   error
}

func (r *memoryNotifications) Save(_ context.Context, n domain.Notification) error {
	if r.err != nil {
		return r.err
	}
	r.items = append(r.items, n)
	return nil
}
func (r *memoryNotifications) ListNotSent(context.Context) ([]domain.Notification, error) {
	return r.items, r.err
}
func (r *memoryNotifications) Mark(_ context.Context, n domain.Notification, s domain.NotificationStatus) error {
	if r.err != nil {
		return r.err
	}
	for i := range r.items {
		if r.items[i].Id == n.Id {
			r.items[i].Status = s
			return nil
		}
	}
	return errors.New("notification not found")
}

// saveSpy is a London-style mock: it records the interaction with the dependency.
type saveSpy struct {
	memoryNotifications
	calls int
	saved domain.Notification
}

func (s *saveSpy) Save(ctx context.Context, n domain.Notification) error {
	s.calls++
	s.saved = n
	return s.memoryNotifications.Save(ctx, n)
}

func TestNotifierUsecase_SaveNotification(t *testing.T) {
	allure.Test(t, "positive classic: stores subscription", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		repo := &memoryNotifications{}
		uc := &NotifierUsecase{repo: repo}
		cmd := subscriptionMother(t)
		cmd.NotifyBeforeMinutes = 30
		// Act
		err := uc.SaveNotification(context.Background(), cmd)
		// Assert
		if err != nil {
			t.Fatal(err)
		}
		wantSendAt := cmd.ScheduledDeparture.Add(-30 * time.Minute)
		if len(repo.items) != 1 || repo.items[0].Type != domain.NotificationSubscribed || !repo.items[0].SendAt.Equal(wantSendAt) {
			t.Fatalf("stored notifications: %+v", repo.items)
		}
		var payload command.SubscriptionCreatedPayloadModel
		if err := json.Unmarshal(repo.items[0].Payload, &payload); err != nil || payload.Email != cmd.Email.String() {
			t.Fatalf("payload: %+v, err: %v", payload, err)
		}
	})
	allure.Test(t, "negative classic: expired notification is skipped", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		repo := &memoryNotifications{}
		uc := &NotifierUsecase{repo: repo}
		cmd := subscriptionMother(t)
		cmd.ScheduledDeparture = time.Now().UTC().Add(-time.Hour)
		// Act
		err := uc.SaveNotification(context.Background(), cmd)
		// Assert
		if err != nil || len(repo.items) != 0 {
			t.Fatalf("err=%v, items=%v", err, repo.items)
		}
	})
	allure.Test(t, "positive London: calls repository once", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		repo := &saveSpy{}
		uc := &NotifierUsecase{repo: repo}
		cmd := subscriptionMother(t)
		// Act
		err := uc.SaveNotification(context.Background(), cmd)
		// Assert
		if err != nil || repo.calls != 1 || repo.saved.Status != domain.NotificationCreated {
			t.Fatalf("err=%v, calls=%d, saved=%+v", err, repo.calls, repo.saved)
		}
	})
	allure.Test(t, "negative London: propagates repository error", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		failure := errors.New("database unavailable")
		repo := &saveSpy{memoryNotifications: memoryNotifications{err: failure}}
		uc := &NotifierUsecase{repo: repo}
		// Act
		err := uc.SaveNotification(context.Background(), subscriptionMother(t))
		// Assert
		if !errors.Is(err, failure) || repo.calls != 1 {
			t.Fatalf("err=%v, calls=%d", err, repo.calls)
		}
	})
}

func TestNotifierUsecase_UpdateFlight(t *testing.T) {
	allure.Test(t, "positive classic: stores urgent update", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		repo := &memoryNotifications{}
		uc := &NotifierUsecase{repo: repo}
		cmd := newUpdateBuilder(t).build()
		// Act
		err := uc.UpdateFlight(context.Background(), cmd)
		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if len(repo.items) != 1 || repo.items[0].Status != domain.NotificationUrgent || repo.items[0].Type != domain.NotificationFlightUpdated {
			t.Fatalf("items=%+v", repo.items)
		}
		var payload command.FlightUpdatedPayloadModel
		if err := json.Unmarshal(repo.items[0].Payload, &payload); err != nil || payload.FlightId != cmd.FlightId {
			t.Fatalf("payload=%+v, err=%v", payload, err)
		}
	})
	allure.Test(t, "negative classic: no recipients", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		repo := &memoryNotifications{}
		uc := &NotifierUsecase{repo: repo}
		cmd := newUpdateBuilder(t).withoutUsers().build()
		// Act
		err := uc.UpdateFlight(context.Background(), cmd)
		// Assert
		if !errors.Is(err, ErrNothingToUpdate) || len(repo.items) != 0 {
			t.Fatalf("err=%v, items=%v", err, repo.items)
		}
	})
	allure.Test(t, "positive London: saves once", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		repo := &saveSpy{}
		uc := &NotifierUsecase{repo: repo}
		// Act
		err := uc.UpdateFlight(context.Background(), newUpdateBuilder(t).build())
		// Assert
		if err != nil || repo.calls != 1 || repo.saved.Status != domain.NotificationUrgent {
			t.Fatalf("err=%v, calls=%d, saved=%+v", err, repo.calls, repo.saved)
		}
	})
	allure.Test(t, "negative London: repository error", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		failure := errors.New("database unavailable")
		repo := &saveSpy{memoryNotifications: memoryNotifications{err: failure}}
		uc := &NotifierUsecase{repo: repo}
		// Act
		err := uc.UpdateFlight(context.Background(), newUpdateBuilder(t).build())
		// Assert
		if !errors.Is(err, failure) || repo.calls != 1 {
			t.Fatalf("err=%v, calls=%d", err, repo.calls)
		}
	})
}

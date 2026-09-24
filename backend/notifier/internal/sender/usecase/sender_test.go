package usecase

import (
	"context"
	"errors"
	receiverCommand "notifier/internal/receiver/command"
	"notifier/internal/receiver/domain"
	"shared/common"
	"testing"
	"time"

	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/google/uuid"
)

// senderRepository is a fixture with real in-memory state and observable calls.
type senderRepository struct {
	items     []domain.Notification
	listErr   error
	markErr   error
	markCalls int
}

type mailSpy struct {
	calls      int
	recipients []common.Email
	message    []byte
	err        error
}

func (m *mailSpy) SendEmail(to []common.Email, msg []byte) error {
	m.calls++
	m.recipients = append([]common.Email(nil), to...)
	m.message = append([]byte(nil), msg...)
	return m.err
}

func (r *senderRepository) Save(_ context.Context, n domain.Notification) error {
	r.items = append(r.items, n)
	return nil
}
func (r *senderRepository) ListNotSent(context.Context) ([]domain.Notification, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.items, nil
}
func (r *senderRepository) Mark(_ context.Context, n domain.Notification, status domain.NotificationStatus) error {
	r.markCalls++
	if r.markErr != nil {
		return r.markErr
	}
	for i := range r.items {
		if r.items[i].Id == n.Id {
			r.items[i].Status = status
			return nil
		}
	}
	return errors.New("notification missing")
}

func senderNotificationMother(t *testing.T) domain.Notification {
	t.Helper()
	payload, err := receiverCommand.SubscriptionCreatedCommandToDomain(
		&receiverCommand.SubscriptionCreatedCommand{
			Email:                 mustEmail(t, "traveler@example.com"),
			ScheduledDeparture:    time.Now().UTC().Add(10 * time.Minute),
			DepartureAirportTitle: "Sheremetyevo", ArrivalAirportTitle: "Pulkovo",
			DepartureAirportCity: mustCity(t), ArrivalAirportCity: mustCity(t),
			DepartureAirportCountry: mustCountry(t), ArrivalAirportCountry: mustCountry(t),
		}, time.Now().UTC(), domain.NotificationCreated)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func updatedNotificationMother(t *testing.T) domain.Notification {
	t.Helper()
	status := "boarding"
	n, err := receiverCommand.FlightUpdatedCommandToDomain(receiverCommand.FlightUpdatedCommand{
		FlightId: uuid.New(), Users: []common.Email{mustEmail(t, "traveler@example.com")},
		DepartureAirportTitle: "Sheremetyevo", ArrivalAirportTitle: "Pulkovo", Status: &status,
	}, time.Now().UTC(), domain.NotificationUrgent)
	if err != nil {
		t.Fatal(err)
	}
	return n
}
func mustEmail(t *testing.T, s string) common.Email {
	t.Helper()
	e, err := common.NewEmail(s)
	if err != nil {
		t.Fatal(err)
	}
	return e
}
func mustCity(t *testing.T) common.City {
	t.Helper()
	c, err := common.NewCity("Moscow")
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func mustCountry(t *testing.T) common.Country {
	t.Helper()
	c, err := common.NewCountry("RU")
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func senderFixture(t *testing.T, repo *senderRepository) *SenderUsecase {
	t.Helper()
	return senderFixtureWithMail(t, repo, &mailSpy{})
}
func senderFixtureWithMail(t *testing.T, repo *senderRepository, m *mailSpy) *SenderUsecase {
	t.Helper()
	return &SenderUsecase{repo: repo, emailSenderUc: &EmailSenderUsecase{m: m, appEmail: mustEmail(t, "service@example.com")}}
}

func TestSenderUsecase_Send(t *testing.T) {
	allure.Test(t, "negative London: mail failure leaves notification pending", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		failure := errors.New("mail transport unavailable")
		repo := &senderRepository{items: []domain.Notification{senderNotificationMother(t)}}
		mail := &mailSpy{err: failure}
		uc := senderFixtureWithMail(t, repo, mail)
		// Act
		err := uc.Send(context.Background())
		// Assert
		if !errors.Is(err, failure) || mail.calls != 1 || repo.markCalls != 0 || repo.items[0].Status != domain.NotificationCreated {
			t.Fatalf("err=%v, mail=%d, mark=%d", err, mail.calls, repo.markCalls)
		}
	})
	allure.Test(t, "positive: urgent update sends immediately", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		n := updatedNotificationMother(t)
		n.SendAt = time.Now().UTC().Add(time.Hour)
		repo := &senderRepository{items: []domain.Notification{n}}
		uc := senderFixture(t, repo)
		// Act
		err := uc.Send(context.Background())
		// Assert
		if err != nil || repo.items[0].Status != domain.NotificationSent || repo.markCalls != 1 {
			t.Fatalf("err=%v, status=%v, calls=%d", err, repo.items[0].Status, repo.markCalls)
		}
	})
	allure.Test(t, "positive classic: sends due notification and marks it", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		repo := &senderRepository{items: []domain.Notification{senderNotificationMother(t)}}
		uc := senderFixture(t, repo)
		// Act
		err := uc.Send(context.Background())
		// Assert
		if err != nil || repo.items[0].Status != domain.NotificationSent || repo.markCalls != 1 {
			t.Fatalf("err=%v, status=%v, markCalls=%d", err, repo.items[0].Status, repo.markCalls)
		}
	})
	allure.Test(t, "negative classic: future notification remains queued", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		n := senderNotificationMother(t)
		n.SendAt = time.Now().UTC().Add(time.Hour)
		repo := &senderRepository{items: []domain.Notification{n}}
		uc := senderFixture(t, repo)
		// Act
		err := uc.Send(context.Background())
		// Assert
		if err != nil || repo.items[0].Status != domain.NotificationCreated || repo.markCalls != 0 {
			t.Fatalf("err=%v, status=%v, markCalls=%d", err, repo.items[0].Status, repo.markCalls)
		}
	})
	allure.Test(t, "negative London: list failure prevents marking", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		failure := errors.New("database unavailable")
		repo := &senderRepository{listErr: failure}
		uc := senderFixture(t, repo)
		// Act
		err := uc.Send(context.Background())
		// Assert
		if !errors.Is(err, failure) || repo.markCalls != 0 {
			t.Fatalf("err=%v, markCalls=%d", err, repo.markCalls)
		}
	})
	allure.Test(t, "negative London: mark failure propagates", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		failure := errors.New("update failed")
		repo := &senderRepository{items: []domain.Notification{senderNotificationMother(t)}, markErr: failure}
		uc := senderFixture(t, repo)
		// Act
		err := uc.Send(context.Background())
		// Assert
		if !errors.Is(err, failure) || repo.markCalls != 1 {
			t.Fatalf("err=%v, markCalls=%d", err, repo.markCalls)
		}
	})
}

func TestEmailSenderUsecase_SendEmail(t *testing.T) {
	allure.Test(t, "positive: valid subscription payload", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		mail := &mailSpy{}
		uc := senderFixtureWithMail(t, &senderRepository{}, mail).emailSenderUc
		n := senderNotificationMother(t)
		// Act
		err := uc.SendEmail(context.Background(), n)
		// Assert
		if err != nil || mail.calls != 1 || len(mail.recipients) != 1 || mail.recipients[0].String() != "traveler@example.com" {
			t.Fatalf("err=%v, mail=%+v", err, mail)
		}
	})
	allure.Test(t, "negative: malformed payload", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		uc := senderFixture(t, &senderRepository{}).emailSenderUc
		n := domain.Notification{Id: uuid.New(), Type: domain.NotificationSubscribed, Payload: []byte("{")}
		// Act
		err := uc.SendEmail(context.Background(), n)
		// Assert
		if err == nil {
			t.Fatal("malformed payload accepted")
		}
	})
	allure.Test(t, "positive: valid flight update payload", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		mail := &mailSpy{}
		uc := senderFixtureWithMail(t, &senderRepository{}, mail).emailSenderUc
		n := updatedNotificationMother(t)
		// Act
		err := uc.SendEmail(context.Background(), n)
		// Assert
		if err != nil || mail.calls != 1 || len(mail.recipients) != 1 {
			t.Fatalf("err=%v, mail=%+v", err, mail)
		}
	})
	allure.Test(t, "negative: flight update with invalid recipient", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		uc := senderFixture(t, &senderRepository{}).emailSenderUc
		n := domain.Notification{Type: domain.NotificationFlightUpdated, Payload: []byte(`{"users":["invalid"]}`)}
		// Act
		err := uc.SendEmail(context.Background(), n)
		// Assert
		if err == nil {
			t.Fatal("invalid recipient accepted")
		}
	})
	allure.Test(t, "negative London: mail transport failure", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		failure := errors.New("mail transport unavailable")
		mail := &mailSpy{err: failure}
		uc := senderFixtureWithMail(t, &senderRepository{}, mail).emailSenderUc
		// Act
		err := uc.SendEmail(context.Background(), senderNotificationMother(t))
		// Assert
		if !errors.Is(err, failure) || mail.calls != 1 {
			t.Fatalf("err=%v, calls=%d", err, mail.calls)
		}
	})
}

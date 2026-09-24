package usecase

import (
	airportDomain "api/internal/airport/domain"
	gateRepository "api/internal/airport/domain/repository"
	flightDomain "api/internal/flight/domain"
	flightRepository "api/internal/flight/domain/repository"
	userDomain "api/internal/user/domain"
	userRepository "api/internal/user/domain/repository"
	"context"
	"encoding/json"
	"errors"
	"shared/common"
	"testing"
	"time"

	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/google/uuid"
)

type userStub struct {
	userRepository.UserRepository
	user  userDomain.User
	err   error
	calls int
}

func (s *userStub) Exist(_ context.Context, _ uuid.UUID) (userDomain.User, error) {
	s.calls++
	return s.user, s.err
}

type flightStub struct {
	flightRepository.FlightRepository
	flight flightDomain.Flight
	route  flightDomain.FlightRoute
	err    error
	calls  int
}

func (s *flightStub) Exist(_ context.Context, _ uuid.UUID) (flightDomain.Flight, error) {
	s.calls++
	return s.flight, s.err
}
func (s *flightStub) GetFlightRoute(_ context.Context, _ uuid.UUID) (flightDomain.FlightRoute, error) {
	return s.route, nil
}

type gateStub struct {
	gateRepository.GateRepository
	airports map[uuid.UUID]airportDomain.Airport
	calls    int
}

func (s *gateStub) GetAirportByGateId(_ context.Context, id uuid.UUID) (airportDomain.Airport, error) {
	s.calls++
	return s.airports[id], nil
}

type senderMock struct {
	topic   string
	message []byte
	calls   int
	err     error
}

func (s *senderMock) SendMessage(topic string, msg []byte) error {
	s.calls++
	s.topic, s.message = topic, append([]byte(nil), msg...)
	return s.err
}

func notificationFixture() (*NotificationUsecase, uuid.UUID, uuid.UUID, *senderMock) {
	uid, fid := uuid.New(), uuid.New()
	depGate, arrGate := uuid.New(), uuid.New()
	sender := &senderMock{}
	uc := &NotificationUsecase{
		ns:       sender,
		userRepo: &userStub{user: userDomain.User{Id: uid, Email: common.Email("traveler@example.com")}},
		flightRepo: &flightStub{
			flight: flightDomain.Flight{Id: fid, ScheduledDeparture: time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC), Status: flightDomain.Scheduled},
			route:  flightDomain.FlightRoute{FlightId: fid, DepartureGateId: depGate, ArrivalGateId: arrGate},
		},
		gateRepo: &gateStub{airports: map[uuid.UUID]airportDomain.Airport{
			depGate: {IATACode: airportDomain.IATACode("SVO"), Title: airportDomain.Title("Sheremetyevo"), City: common.City("Moscow"), Country: common.Country("RU")},
			arrGate: {IATACode: airportDomain.IATACode("LED"), Title: airportDomain.Title("Pulkovo"), City: common.City("Saint Petersburg"), Country: common.Country("RU")},
		}},
	}
	return uc, uid, fid, sender
}

func TestNotificationUsecase_SendMessage(t *testing.T) {
	allure.Test(t, "positive: sends subscription notification", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		uc, uid, fid, sender := notificationFixture()
		// Act
		err := uc.SendMessage(uid, fid)
		// Assert
		if err != nil || sender.calls != 1 || sender.topic != "flights" {
			t.Fatalf("err=%v, sender=%+v", err, sender)
		}
		var payload map[string]any
		if err := json.Unmarshal(sender.message, &payload); err != nil || payload["email"] != "traveler@example.com" {
			t.Fatalf("payload=%+v, err=%v", payload, err)
		}
	})
	allure.Test(t, "negative: unknown user stops workflow", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		uc, uid, fid, sender := notificationFixture()
		uc.userRepo.(*userStub).err = userRepository.ErrUserNotFound
		// Act
		err := uc.SendMessage(uid, fid)
		// Assert
		if !errors.Is(err, ErrUserNotFound) || sender.calls != 0 || uc.flightRepo.(*flightStub).calls != 0 {
			t.Fatalf("err=%v, sender=%+v", err, sender)
		}
	})
	allure.Test(t, "negative: unknown flight stops workflow", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		uc, uid, fid, sender := notificationFixture()
		uc.flightRepo.(*flightStub).err = flightRepository.ErrFlightNotFound
		// Act
		err := uc.SendMessage(uid, fid)
		// Assert
		if !errors.Is(err, ErrFlightNotFound) || sender.calls != 0 || uc.gateRepo.(*gateStub).calls != 0 {
			t.Fatalf("err=%v, sender=%+v", err, sender)
		}
	})
	allure.Test(t, "negative: broker error propagates", func(allureContext *allure.Context) {
		t := allureContext.T()

		// Arrange
		uc, uid, fid, sender := notificationFixture()
		failure := errors.New("broker unavailable")
		sender.err = failure
		// Act
		err := uc.SendMessage(uid, fid)
		// Assert
		if !errors.Is(err, failure) || sender.calls != 1 {
			t.Fatalf("err=%v, sender=%+v", err, sender)
		}
	})
}

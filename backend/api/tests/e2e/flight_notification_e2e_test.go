//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	allure "github.com/allure-framework/allure-go/commons/gotest"
	"github.com/google/uuid"
)

type flightInfo struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Plan   string `json:"plan"`
}

type message struct {
	ID      string
	Subject string
	Text    string
	To      []struct{ Address string }
}

func TestFlightSubscriptionNotification(t *testing.T) {
	allure.Test(t, "e2e: user receives an email after an administrator changes ctx subscribed flight", func(ctx *allure.Context) {
		api, mailpit := os.Getenv("API_BASE_URL"), os.Getenv("MAILPIT_URL")
		if api == "" || mailpit == "" {
			ctx.T().Fatal("API_BASE_URL and MAILPIT_URL are required")
		}
		const flightID = "00000000-0000-4000-8000-000000000050"
		email := "passenger-" + uuid.NewString() + "@airline.test"
		const password = "Passenger123!"
		plan := "E2E " + strings.ToUpper(uuid.NewString())
		var passengerToken, adminToken string

		ctx.Step("Register ctx passenger and log in", func(ctx *allure.Context) {
			credentials := map[string]string{"email": email, "password": password}
			requestJSON(ctx, "POST", api+"/users", "", credentials, http.StatusCreated, nil)
			var login struct{ Token string }
			requestJSON(ctx, "POST", api+"/auth/tokens", "", credentials, http.StatusOK, &login)
			if login.Token == "" {
				ctx.T().Fatal("passenger did not receive ctx token")
			}
			passengerToken = login.Token
		})

		ctx.Step("Find the scheduled flight", func(ctx *allure.Context) {
			var catalogue struct{ Items []flightInfo }
			requestJSON(ctx, "GET", api+"/flights?limit=10", "", nil, http.StatusOK, &catalogue)
			for _, flight := range catalogue.Items {
				if flight.ID == flightID && flight.Status == "scheduled" {
					return
				}
			}
			ctx.T().Fatal("the scheduled flight is missing")
		})

		ctx.Step("Subscribe and check the passenger's subscriptions", func(ctx *allure.Context) {
			requestJSON(ctx, "POST", api+"/subscriptions", passengerToken,
				map[string]any{"flight_id": flightID, "notify_before_minutes": 60}, http.StatusOK, nil)
			var subscriptions struct{ Items []flightInfo }
			requestJSON(ctx, "GET", api+"/subscriptions", passengerToken, nil, http.StatusOK, &subscriptions)
			if len(subscriptions.Items) != 1 || subscriptions.Items[0].ID != flightID {
				ctx.T().Fatalf("unexpected subscriptions: %+v", subscriptions.Items)
			}
		})

		ctx.Step("Log in as administrator and change the flight", func(ctx *allure.Context) {
			var login struct{ Token string }
			requestJSON(ctx, "POST", api+"/auth/tokens", "",
				map[string]string{"email": "admin@airline.test", "password": "TestAdmin123!"}, http.StatusOK, &login)
			if login.Token == "" {
				ctx.T().Fatal("administrator did not receive ctx token")
			}
			adminToken = login.Token
			requestJSON(ctx, "PATCH", api+"/flights/"+flightID, adminToken,
				map[string]any{"flight": map[string]string{"status": "delayed", "plan": plan}}, http.StatusOK, nil)
		})

		ctx.Step("Check the public flight details", func(ctx *allure.Context) {
			var response struct{ Flight flightInfo }
			requestJSON(ctx, "GET", api+"/flights/"+flightID, "", nil, http.StatusOK, &response)
			if response.Flight.ID != flightID || response.Flight.Status != "delayed" || response.Flight.Plan != plan {
				ctx.T().Fatalf("flight changes were not published: %+v", response.Flight)
			}
		})

		ctx.Step("Receive the flight change email in Mailpit", func(ctx *allure.Context) {
			mail := waitForEmail(ctx, mailpit, email, plan)
			if mail.Subject != "Flight departure" ||
				!strings.Contains(mail.Text, "from Sheremetyevo to Pulkovo has changed") ||
				!strings.Contains(mail.Text, "New flight's status is delayed") {
				ctx.T().Fatalf("unexpected notification: subject=%q, body=%q", mail.Subject, mail.Text)
			}
		})
	})
}

func requestJSON(ctx *allure.Context, method, url, token string, input any, wantStatus int, output any) {
	ctx.T().Helper()
	var payload []byte
	var err error
	if input != nil {
		payload, err = json.Marshal(input)
		if err != nil {
			ctx.T().Fatal(err)
		}
	}
	request, err := http.NewRequestWithContext(ctx.T().Context(), method, url, bytes.NewReader(payload))
	if err != nil {
		ctx.T().Fatal(err)
	}
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		ctx.T().Fatalf("%s %s: %v", method, url, err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		ctx.T().Fatal(err)
	}
	ctx.Attachment(method+" "+url, []byte(fmt.Sprintf("Request: %s\nStatus: %d\nResponse: %s", payload, response.StatusCode, body)), "text/plain")
	if response.StatusCode != wantStatus {
		ctx.T().Fatalf("%s %s: status=%d, want=%d, body=%s", method, url, response.StatusCode, wantStatus, body)
	}
	if output != nil {
		if err := json.Unmarshal(body, output); err != nil {
			ctx.T().Fatalf("decode %s: %v", url, err)
		}
	}
}

func waitForEmail(ctx *allure.Context, mailpit, recipient, marker string) message {
	ctx.T().Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		var mailbox struct{ Messages []message }
		requestJSON(ctx, "GET", mailpit+"/api/v1/messages?limit=50", "", nil, http.StatusOK, &mailbox)
		for _, summary := range mailbox.Messages {
			for _, to := range summary.To {
				if to.Address != recipient {
					continue
				}
				var mail message
				requestJSON(ctx, "GET", mailpit+"/api/v1/message/"+summary.ID, "", nil, http.StatusOK, &mail)
				if strings.Contains(mail.Text, marker) {
					return mail
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	ctx.T().Fatalf("no flight change email for %s with marker %q within 30 seconds", recipient, marker)
	return message{}
}

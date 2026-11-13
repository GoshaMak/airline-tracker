//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

//go:embed features/authentication.feature
var authFeatureFS embed.FS

type authBDD struct {
	api         string
	mailpit     string
	postgres    string
	email       string
	password    string
	newPassword string
	challengeID string
	token       string
}

func TestAuthenticationBDD(t *testing.T) {
	options := godog.Options{
		Format:   "pretty",
		Paths:    []string{"features/authentication.feature"},
		FS:       authFeatureFS,
		TestingT: t,
	}
	suite := godog.TestSuite{
		ScenarioInitializer: initializeAuthScenario,
		Options:             &options,
	}
	if suite.Run() != 0 {
		t.Fail()
	}
}

func initializeAuthScenario(sc *godog.ScenarioContext) {
	state := &authBDD{
		api:      strings.TrimRight(os.Getenv("API_BASE_URL"), "/"),
		mailpit:  strings.TrimRight(os.Getenv("MAILPIT_URL"), "/"),
		postgres: os.Getenv("POSTGRES_URI"),
	}
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		if state.api == "" || state.mailpit == "" || state.postgres == "" {
			return ctx, errors.New("API_BASE_URL, MAILPIT_URL and POSTGRES_URI are required")
		}
		state.email = "auth-e2e-" + uuid.NewString() + "@airline.test"
		state.password = generatedPassword()
		state.newPassword = generatedPassword()
		state.challengeID, state.token = "", ""
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, scenarioErr error) (context.Context, error) {
		if state.email == "" {
			return ctx, nil
		}
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		conn, err := pgx.Connect(cleanupCtx, state.postgres)
		if err != nil {
			return ctx, fmt.Errorf("connect to test database to remove fixture user: %w", err)
		}
		defer conn.Close(cleanupCtx)
		if _, err := conn.Exec(cleanupCtx, "delete from users where email = $1", state.email); err != nil {
			return ctx, fmt.Errorf("remove fixture user %s: %w", state.email, err)
		}
		return ctx, nil
	})

	sc.Step(`^a registered technical user$`, state.registerUser)
	sc.Step(`^I submit the correct password$`, state.submitCorrectPassword)
	sc.Step(`^the API asks for email verification$`, state.assertChallenge)
	sc.Step(`^I submit the code from the verification email$`, state.submitVerificationCode)
	sc.Step(`^I receive an authenticated session$`, state.assertAuthenticated)
	sc.Step(`^I submit an incorrect password five times$`, state.submitWrongPasswordFiveTimes)
	sc.Step(`^the account is locked$`, state.assertLocked)
	sc.Step(`^I request account recovery$`, state.requestRecovery)
	sc.Step(`^I submit the code from the recovery email$`, state.submitRecoveryCode)
	sc.Step(`^the account is unlocked$`, state.assertUnlocked)
	sc.Step(`^I submit an incorrect verification code five times$`, state.submitWrongCodeFiveTimes)
	sc.Step(`^I change my password$`, state.changePassword)
	sc.Step(`^the new password works and the old password is rejected$`, state.assertPasswordRotation)
}

func (s *authBDD) registerUser(ctx context.Context) error {
	return s.request(ctx,
		http.MethodPost,
		"/users",
		"",
		map[string]string{
			"email":    s.email,
			"password": s.password,
		},
		http.StatusCreated,
		nil,
	)
}

func (s *authBDD) submitCorrectPassword(ctx context.Context) error {
	if err := s.clearMailbox(ctx); err != nil {
		return err
	}
	var response struct {
		ChallengeID string `json:"challenge_id"`
	}
	if err := s.request(ctx,
		http.MethodPost,
		"/auth/tokens",
		"",
		map[string]string{
			"email":    s.email,
			"password": s.password,
		},
		http.StatusAccepted,
		&response); err != nil {
		return err
	}
	if response.ChallengeID == "" {
		return errors.New("login response has an empty challenge_id")
	}
	s.challengeID = response.ChallengeID
	return nil
}

func (s *authBDD) assertChallenge() error {
	if s.challengeID == "" {
		return errors.New("no email verification challenge was created")
	}
	return nil
}

func (s *authBDD) submitVerificationCode(ctx context.Context) error {
	code, err := s.mailCode(ctx, "Airline Tracker login code", `verification code:\s*([0-9]{6})`)
	if err != nil {
		return err
	}
	var response struct {
		Token string `json:"token"`
	}
	if err := s.request(ctx,
		http.MethodPost,
		"/auth/verify",
		"",
		map[string]string{
			"challenge_id": s.challengeID,
			"code":         code,
		},
		http.StatusOK,
		&response); err != nil {
		return err
	}
	s.token = response.Token
	return nil
}

func (s *authBDD) assertAuthenticated(ctx context.Context) error {
	if s.token == "" {
		return errors.New("verification did not issue a session token")
	}
	return s.request(ctx, http.MethodGet, "/subscriptions", s.token, nil, http.StatusOK, nil)
}

func (s *authBDD) submitWrongPasswordFiveTimes(ctx context.Context) error {
	for attempt := 1; attempt <= 5; attempt++ {
		want := http.StatusUnauthorized
		if attempt == 5 {
			want = http.StatusLocked
		}
		if err := s.request(ctx,
			http.MethodPost,
			"/auth/tokens",
			"",
			map[string]string{
				"email":    s.email,
				"password": "Wrong" + uuid.NewString() + "!aA1",
			},
			want,
			nil); err != nil {
			return fmt.Errorf("password attempt %d: %w", attempt, err)
		}
	}
	return nil
}

func (s *authBDD) assertLocked(ctx context.Context) error {
	return s.request(ctx,
		http.MethodPost,
		"/auth/tokens",
		"",
		map[string]string{
			"email":    s.email,
			"password": s.password,
		},
		http.StatusLocked,
		nil)
}

func (s *authBDD) requestRecovery(ctx context.Context) error {
	return s.request(ctx,
		http.MethodPost,
		"/auth/recovery",
		"",
		map[string]string{
			"email": s.email,
		},
		http.StatusAccepted,
		nil)
}

func (s *authBDD) submitRecoveryCode(ctx context.Context) error {
	code, err := s.mailCode(ctx, "Airline Tracker recovery code", `recovery code:\s*([0-9]{6})`)
	if err != nil {
		return err
	}
	if err := s.request(ctx,
		http.MethodPost,
		"/auth/recovery/verify",
		"",
		map[string]string{
			"email": s.email,
			"code":  code,
		},
		http.StatusOK,
		nil); err != nil {
		return err
	}
	return nil
}

func (s *authBDD) assertUnlocked(ctx context.Context) error {
	var response struct {
		ChallengeID string `json:"challenge_id"`
	}
	if err := s.request(ctx,
		http.MethodPost,
		"/auth/tokens",
		"",
		map[string]string{
			"email":    s.email,
			"password": s.password,
		},
		http.StatusAccepted,
		&response); err != nil {
		return fmt.Errorf("account should accept a password after recovery: %w", err)
	}
	if response.ChallengeID == "" {
		return errors.New("unlocked login did not create an email challenge")
	}
	s.challengeID = response.ChallengeID
	return nil
}

func (s *authBDD) submitWrongCodeFiveTimes(ctx context.Context) error {
	validCode, err := s.mailCode(ctx, "Airline Tracker login code", `verification code:\s*([0-9]{6})`)
	if err != nil {
		return err
	}
	wrongCode := "000000"
	if wrongCode == validCode {
		wrongCode = "000001"
	}
	for attempt := 1; attempt <= 5; attempt++ {
		want := http.StatusUnauthorized
		if attempt == 5 {
			want = http.StatusLocked
		}
		if err := s.request(ctx,
			http.MethodPost,
			"/auth/verify",
			"",
			map[string]string{
				"challenge_id": s.challengeID,
				"code":         wrongCode,
			},
			want,
			nil); err != nil {
			return fmt.Errorf("verification attempt %d: %w", attempt, err)
		}
	}
	return nil
}

func (s *authBDD) changePassword(ctx context.Context) error {
	return s.request(ctx,
		http.MethodPost,
		"/auth/password",
		s.token,
		map[string]string{
			"current_password": s.password,
			"new_password":     s.newPassword,
		},
		http.StatusNoContent,
		nil)
}

func (s *authBDD) assertPasswordRotation(ctx context.Context) error {
	var oldResponse map[string]any
	if err := s.request(ctx,
		http.MethodPost,
		"/auth/tokens",
		"",
		map[string]string{
			"email":    s.email,
			"password": s.password,
		},
		http.StatusUnauthorized,
		&oldResponse); err != nil {
		return fmt.Errorf("old password should be rejected: %w", err)
	}
	if err := s.clearMailbox(ctx); err != nil {
		return err
	}
	var response struct {
		ChallengeID string `json:"challenge_id"`
	}
	if err := s.request(ctx,
		http.MethodPost,
		"/auth/tokens",
		"",
		map[string]string{
			"email":    s.email,
			"password": s.newPassword,
		},
		http.StatusAccepted,
		&response); err != nil {
		return fmt.Errorf("new password should be accepted: %w", err)
	}
	s.challengeID = response.ChallengeID
	return s.submitVerificationCode(ctx)
}

func (s *authBDD) mailCode(ctx context.Context, subject, pattern string) (string, error) {
	deadline := time.Now().Add(30 * time.Second)
	codePattern := regexp.MustCompile(pattern)
	for time.Now().Before(deadline) {
		var mailbox struct {
			Messages []struct {
				ID      string `json:"ID"`
				Subject string `json:"Subject"`
				To      []struct {
					Address string `json:"Address"`
				} `json:"To"`
			} `json:"messages"`
		}
		if err := s.requestURL(ctx,
			http.MethodGet,
			s.mailpit+"/api/v1/messages?limit=50",
			"",
			nil,
			http.StatusOK,
			&mailbox); err != nil {
			return "", err
		}
		for _, summary := range mailbox.Messages {
			if summary.Subject != subject {
				continue
			}
			for _, recipient := range summary.To {
				if !strings.EqualFold(recipient.Address, s.email) {
					continue
				}
				var message struct {
					Text string `json:"Text"`
				}
				if err := s.requestURL(ctx,
					http.MethodGet,
					s.mailpit+"/api/v1/message/"+url.PathEscape(summary.ID),
					"",
					nil,
					http.StatusOK,
					&message); err != nil {
					return "", err
				}
				match := codePattern.FindStringSubmatch(message.Text)
				if len(match) == 2 {
					return match[1], nil
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return "", fmt.Errorf("no %q email code for %s within 30 seconds", subject, s.email)
}

func (s *authBDD) clearMailbox(ctx context.Context) error {
	return s.requestURL(ctx, http.MethodDelete, s.mailpit+"/api/v1/messages", "", nil, http.StatusOK, nil)
}

func (s *authBDD) request(ctx context.Context, method, path, token string, input any, status int, output any) error {
	return s.doRequest(ctx, method, s.api+path, token, input, output, status)
}

func (s *authBDD) requestURL(ctx context.Context, method, endpoint, token string, input any, status int, output any) error {
	return s.doRequest(ctx, method, endpoint, token, input, output, status)
}

func (s *authBDD) doRequest(ctx context.Context, method, endpoint, token string, input, output any, wantStatus int) error {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, endpoint, err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode != wantStatus {
		return fmt.Errorf("%s %s: status=%d, want=%d, body=%s", method, endpoint, response.StatusCode, wantStatus, responseBody)
	}
	if output != nil && len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, output); err != nil {
			return fmt.Errorf("decode %s response: %w", endpoint, err)
		}
	}
	return nil
}

func generatedPassword() string {
	return "Aa1!" + strings.ReplaceAll(uuid.NewString(), "-", "")
}

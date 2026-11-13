package redis

import (
	"api/internal/auth/usecase"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"shared/common"
	"time"

	rds "github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
)

//go:embed scripts/record_failed_attempt.lua
var recordFailedAttemptScript string

//go:embed scripts/verify_login_challenge.lua
var verifyLoginChallengeScript string

//go:embed scripts/verify_recovery_code.lua
var verifyRecoveryCodeScript string

type authStateStore struct {
	client *rds.Client
}

func NewAuthStateStore(i do.Injector) (usecase.AuthStateStore, error) {
	return &authStateStore{
		client: do.MustInvoke[*rds.Client](i),
	}, nil
}

func (s *authStateStore) IsLocked(ctx context.Context, email common.Email) (bool, error) {
	count, err := s.client.Exists(ctx, lockKey(email)).Result()
	return count > 0, err
}

func (s *authStateStore) RecordFailedAttempt(
	ctx context.Context,
	email common.Email,
	policy usecase.AuthAttemptPolicy,
) (bool, error) {
	locked, err := s.client.Eval(ctx, recordFailedAttemptScript,
		[]string{
			attemptKey(email),
			lockKey(email),
		},
		policy.MaxAttempts,
		durationSeconds(policy.AttemptWindow),
		durationSeconds(policy.LockDuration),
	).Int64()
	if err != nil {
		return false, fmt.Errorf("redis record failed auth attempt: %w", err)
	}
	return locked == 1, nil
}

func (s *authStateStore) SaveLoginChallenge(
	ctx context.Context,
	id string,
	email common.Email,
	code string,
	ttl time.Duration,
) error {
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, challengeKey(id), email.String(), ttl)
	pipe.Set(ctx, challengeCodeKey(id), digestCode(code), ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis save login challenge: %w", err)
	}
	return nil
}

func (s *authStateStore) DeleteLoginChallenge(ctx context.Context, id string) error {
	return s.client.Del(ctx, challengeKey(id), challengeCodeKey(id)).Err()
}

func (s *authStateStore) VerifyLoginChallenge(
	ctx context.Context,
	id,
	code string,
	policy usecase.AuthAttemptPolicy,
) (common.Email, usecase.AuthStateResult, error) {
	challenge := challengeKey(id)
	storedEmail, err := s.client.Get(ctx, challenge).Result()
	if err == rds.Nil {
		return "", usecase.AuthStateInvalid, nil
	}
	if err != nil {
		return "", usecase.AuthStateInvalid, fmt.Errorf("redis load login challenge: %w", err)
	}
	email, err := common.NewEmail(storedEmail)
	if err != nil {
		return "", usecase.AuthStateInvalid, fmt.Errorf("redis challenge contains invalid email: %w", err)
	}
	result, err := s.client.Eval(ctx, verifyLoginChallengeScript,
		[]string{
			challenge,
			challengeCodeKey(id),
			lockKey(email),
			attemptKey(email),
		},
		digestCode(code),
		policy.MaxAttempts,
		durationSeconds(policy.AttemptWindow),
		durationSeconds(policy.LockDuration),
	).Int64()
	if err != nil {
		return "", usecase.AuthStateInvalid, fmt.Errorf("redis verify login challenge: %w", err)
	}
	switch result {
	case 0:
		return "", usecase.AuthStateInvalid, nil
	case -2:
		return email, usecase.AuthStateRejected, nil
	case -3:
		return email, usecase.AuthStateLocked, nil
	case 1:
		return email, usecase.AuthStateAccepted, nil
	default:
		return "", usecase.AuthStateInvalid, fmt.Errorf("redis verify login challenge returned unexpected result %d", result)
	}
}

func (s *authStateStore) SaveRecoveryCode(
	ctx context.Context,
	email common.Email,
	code string,
	ttl time.Duration,
) error {
	return s.client.Set(ctx, recoveryKey(email), digestCode(code), ttl).Err()
}

func (s *authStateStore) DeleteRecoveryCode(ctx context.Context, email common.Email) error {
	return s.client.Del(ctx, recoveryKey(email)).Err()
}

func (s *authStateStore) VerifyRecoveryCode(
	ctx context.Context,
	email common.Email,
	code string,
	policy usecase.AuthAttemptPolicy,
) (bool, error) {
	result, err := s.client.Eval(ctx, verifyRecoveryCodeScript,
		[]string{
			recoveryKey(email),
			lockKey(email),
			attemptKey(email),
			recoveryAttemptKey(email),
		},
		digestCode(code),
		policy.MaxAttempts,
		durationSeconds(policy.AttemptWindow),
	).Int64()
	if err != nil {
		return false, fmt.Errorf("redis verify recovery code: %w", err)
	}
	return result == 1, nil
}

func challengeKey(id string) string {
	return "auth:challenge:" + id
}

func challengeCodeKey(id string) string {
	return "auth:challenge-code:" + id
}

func recoveryKey(email common.Email) string {
	return "auth:recovery:" + emailKey(email)
}

func attemptKey(email common.Email) string {
	return "auth:attempts:" + emailKey(email)
}

func lockKey(email common.Email) string {
	return "auth:locked:" + emailKey(email)
}

func recoveryAttemptKey(email common.Email) string {
	return "auth:recovery-attempts:" + emailKey(email)
}

func emailKey(email common.Email) string {
	hash := sha256.Sum256([]byte(email.String()))
	return hex.EncodeToString(hash[:])
}

func digestCode(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}

func durationSeconds(duration time.Duration) int64 {
	seconds := int64(duration / time.Second)
	if duration%time.Second != 0 {
		seconds++
	}
	return seconds
}

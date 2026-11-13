package usecase

import (
	"context"
	"shared/common"
	"time"
)

type AuthAttemptPolicy struct {
	MaxAttempts   int
	AttemptWindow time.Duration
	LockDuration  time.Duration
}

type AuthStateResult int

const (
	AuthStateInvalid AuthStateResult = iota
	AuthStateRejected
	AuthStateLocked
	AuthStateAccepted
)

type AuthStateStore interface {
	IsLocked(ctx context.Context, email common.Email) (bool, error)
	RecordFailedAttempt(ctx context.Context, email common.Email, policy AuthAttemptPolicy) (bool, error)
	SaveLoginChallenge(ctx context.Context, id string, email common.Email, code string, ttl time.Duration) error
	DeleteLoginChallenge(ctx context.Context, id string) error
	VerifyLoginChallenge(ctx context.Context, id, code string, policy AuthAttemptPolicy) (common.Email, AuthStateResult, error)
	SaveRecoveryCode(ctx context.Context, email common.Email, code string, ttl time.Duration) error
	DeleteRecoveryCode(ctx context.Context, email common.Email) error
	VerifyRecoveryCode(ctx context.Context, email common.Email, code string, policy AuthAttemptPolicy) (bool, error)
}

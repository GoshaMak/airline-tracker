package usecase

import (
	"api/internal/user/domain"
	"api/internal/user/domain/repository"
	"api/pkg"
	"shared/common"

	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"
)

const codeTTL = 5 * time.Minute

var authAttemptPolicy = AuthAttemptPolicy{
	MaxAttempts:   5,
	AttemptWindow: time.Hour,
	LockDuration:  15 * time.Minute,
}

var recoveryAttemptPolicy = AuthAttemptPolicy{
	MaxAttempts:   5,
	AttemptWindow: 15 * time.Minute,
	LockDuration:  15 * time.Minute,
}

type AuthUsecase struct {
	repo   repository.UserRepository
	state  AuthStateStore
	mailer pkg.Mailer
}

func NewAuthUsecase(i do.Injector) (*AuthUsecase, error) {
	mailer, err := pkg.NewMailingClient(
		os.Getenv("SMTP_ADDR"),
		os.Getenv("APP_EMAIL"),
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
	)
	if err != nil {
		return nil, fmt.Errorf("configure auth mailing client: %w", err)
	}
	return &AuthUsecase{
		repo:   do.MustInvoke[repository.UserRepository](i),
		state:  do.MustInvoke[AuthStateStore](i),
		mailer: mailer,
	}, nil
}

func (uc *AuthUsecase) GetUser(ctx context.Context, email common.Email, password string) (domain.User, error) {
	u, err := uc.repo.GetUser(ctx, email.String())
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return domain.User{}, ErrUserNotFound
		}
		if errors.Is(err, repository.ErrOperationNotPermitted) {
			return domain.User{}, ErrOperationNotPermitted
		}
		return domain.User{}, fmt.Errorf("AuthUsecase.GetUser: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return domain.User{}, ErrInvalidPassword
	}
	return u, nil
}

func (uc *AuthUsecase) BeginLogin(ctx context.Context, email common.Email, password string) (string, error) {
	locked, err := uc.state.IsLocked(ctx, email)
	if err != nil {
		return "", fmt.Errorf("check account lock: %w", err)
	}
	if locked {
		return "", ErrAccountLocked
	}
	u, err := uc.repo.GetUser(ctx, email.String())
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", uc.failedAttempt(ctx, email)
		}
		return "", fmt.Errorf("AuthUsecase.BeginLogin: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", uc.failedAttempt(ctx, email)
	}

	code, err := randomCode()
	if err != nil {
		return "", fmt.Errorf("generate login code: %w", err)
	}
	challengeID := uuid.NewString()
	if err := uc.state.SaveLoginChallenge(ctx, challengeID, email, code, codeTTL); err != nil {
		return "", fmt.Errorf("save login challenge: %w", err)
	}
	if err := uc.sendCodeEmail(ctx, email, "Airline Tracker login code", "verification code", code); err != nil {
		if deleteErr := uc.state.DeleteLoginChallenge(ctx, challengeID); deleteErr != nil {
			return "", fmt.Errorf("send login code: %w (cleanup challenge: %v)", err, deleteErr)
		}
		return "", fmt.Errorf("send login code: %w", err)
	}
	return challengeID, nil
}

func (uc *AuthUsecase) VerifyLogin(ctx context.Context, challengeID, code string) (string, error) {
	email, result, err := uc.state.VerifyLoginChallenge(ctx, challengeID, code, authAttemptPolicy)
	if err != nil {
		return "", fmt.Errorf("verify login challenge: %w", err)
	}
	switch result {
	case AuthStateLocked:
		return "", ErrAccountLocked
	case AuthStateRejected:
		return "", ErrInvalidPassword
	case AuthStateInvalid:
		return "", ErrInvalidCode
	case AuthStateAccepted:
	default:
		return "", fmt.Errorf("unexpected login verification result %d", result)
	}
	u, err := uc.repo.GetUser(ctx, email.String())
	if err != nil {
		return "", fmt.Errorf("reload user after login challenge: %w", err)
	}
	return GenerateJWT(u)
}

func (uc *AuthUsecase) RequestRecovery(ctx context.Context, email common.Email) error {
	locked, err := uc.state.IsLocked(ctx, email)
	if err != nil {
		return fmt.Errorf("check account lock: %w", err)
	}
	if !locked {
		return nil
	}
	if _, err := uc.repo.GetUser(ctx, email.String()); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil
		}
		return fmt.Errorf("load recovery user: %w", err)
	}
	code, err := randomCode()
	if err != nil {
		return fmt.Errorf("generate recovery code: %w", err)
	}
	if err := uc.state.SaveRecoveryCode(ctx, email, code, recoveryAttemptPolicy.LockDuration); err != nil {
		return fmt.Errorf("save recovery code: %w", err)
	}
	if err := uc.sendCodeEmail(ctx, email, "Airline Tracker recovery code", "recovery code", code); err != nil {
		if deleteErr := uc.state.DeleteRecoveryCode(ctx, email); deleteErr != nil {
			return fmt.Errorf("send recovery code: %w (cleanup recovery code: %v)", err, deleteErr)
		}
		return fmt.Errorf("send recovery code: %w", err)
	}
	return nil
}

func (uc *AuthUsecase) VerifyRecovery(ctx context.Context, email common.Email, code string) error {
	verified, err := uc.state.VerifyRecoveryCode(ctx, email, code, recoveryAttemptPolicy)
	if err != nil {
		return fmt.Errorf("verify recovery code: %w", err)
	}
	if !verified {
		return ErrInvalidCode
	}
	return nil
}

func (uc *AuthUsecase) ChangePassword(ctx context.Context, uid uuid.UUID, currentPassword, newPassword string) error {
	u, err := uc.repo.Exist(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("load user for password change: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(currentPassword)) != nil {
		return ErrInvalidPassword
	}
	if err := domain.IsValidPassword(newPassword); err != nil {
		return ErrWeakPassword
	}
	passwordHash, err := domain.NewPasswordHashed(newPassword)
	if err != nil {
		return err
	}
	if err := uc.repo.UpdatePassword(ctx, uid, passwordHash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func (uc *AuthUsecase) failedAttempt(ctx context.Context, email common.Email) error {
	locked, err := uc.state.RecordFailedAttempt(ctx, email, authAttemptPolicy)
	if err != nil {
		return fmt.Errorf("record authentication attempt: %w", err)
	}
	if locked {
		return ErrAccountLocked
	}
	return ErrInvalidPassword
}

func randomCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Uint64()), nil
}

func (uc *AuthUsecase) sendCodeEmail(ctx context.Context, recipient common.Email, subject, label, code string) error {
	body := fmt.Sprintf("Your %s: %s\r\nIt expires in %d minutes.", label, code, int(codeTTL.Minutes()))
	return uc.mailer.SendEmail(ctx, recipient, subject, body)
}

func (uc *AuthUsecase) CreateUser(ctx context.Context, u domain.User) error {
	if err := uc.repo.SaveUser(ctx, u); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return ErrUserAlreadyExists
		}
		if errors.Is(err, repository.ErrOperationNotPermitted) {
			return ErrOperationNotPermitted
		}
		return fmt.Errorf("AuthUsecase.CreateUser: %w", err)
	}
	return nil
}

// TODO: change to (exists bool, err error)
func (uc *AuthUsecase) Exists(ctx context.Context, email common.Email, password string) bool {
	_, err := uc.GetUser(ctx, email, password)
	return err == nil
}

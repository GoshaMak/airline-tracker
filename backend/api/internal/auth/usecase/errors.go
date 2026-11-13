package usecase

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrOperationNotPermitted = errors.New("operation not permitted")
	ErrAccountLocked         = errors.New("account is locked")
	ErrInvalidCode           = errors.New("invalid or expired code")
	ErrWeakPassword          = errors.New("new password does not meet policy")
)

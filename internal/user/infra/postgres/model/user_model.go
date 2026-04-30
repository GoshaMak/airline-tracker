package model

import (
	"github.com/google/uuid"
)

type UserModel struct {
	ID       uuid.UUID `db:"id"`
	Email    string    `db:"email"`
	Password string    `db:"password_hash"`
	Role     string    `db:"role"`
}

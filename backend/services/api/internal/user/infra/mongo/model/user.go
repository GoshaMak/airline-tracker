package model

import (
	"api/internal/user/domain"
	"shared/common"

	"github.com/google/uuid"
)

type UserDoc struct {
	ID           string `bson:"_id"`
	Email        string `bson:"email"`
	PasswordHash string `bson:"password_hash"`
	Role         string `bson:"role"`
}

func ToUserDoc(u domain.User) UserDoc {
	return UserDoc{
		ID:           u.Id.String(),
		Email:        u.Email.String(),
		PasswordHash: u.PasswordHash.String(),
		Role:         u.Role.String(),
	}
}

func FromUserDoc(doc UserDoc) (domain.User, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return domain.User{}, err
	}

	email, err := common.NewEmail(doc.Email)
	if err != nil {
		return domain.User{}, err
	}

	password, err := domain.NewPasswordHashed(doc.PasswordHash)
	if err != nil {
		return domain.User{}, err
	}

	role, err := domain.NewRole(doc.Role)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:           id,
		Email:        email,
		PasswordHash: password,
		Role:         role,
	}, nil
}

type SubscriptionDoc struct {
	UserID   string `bson:"user_id"`
	FlightID string `bson:"flight_id"`
}

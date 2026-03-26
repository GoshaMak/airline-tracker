package service

import (
	"airline-tracker/internal/user/domain"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtKey = []byte("")

type JWTClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func newJWTClaims(userID, role string) *JWTClaims {
	return &JWTClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth-service",
			Subject:   userID,
			Audience:  []string{"service"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
	}
}

func GetJWTKey() []byte {
	if len(jwtKey) == 0 {
		jwtKey = []byte(os.Getenv("JWT_KEY"))
	}
	return jwtKey
}

func GenerateJWT(u *domain.User) (string, error) {
	claims := newJWTClaims(u.ID.String(), u.Role)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, *claims)
	key := GetJWTKey()
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

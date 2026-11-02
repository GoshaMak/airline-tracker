package usecase

import (
	"api/internal/user/domain"
	userDomain "api/internal/user/domain"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtOnce sync.Once
var jwtKey = []byte("")

type JWTClaims struct {
	Role userDomain.Role `json:"role"`
	jwt.RegisteredClaims
}

func newJWTClaims(userID string, role userDomain.Role) *JWTClaims {
	now := time.Now()
	return &JWTClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth-service",
			Subject:   userID,
			Audience:  []string{"service"},
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}
}

func GetJWTKey() []byte {
	jwtOnce.Do(func() {
		jwtKey = []byte(os.Getenv("JWT_KEY"))
	})
	return jwtKey
}

func GenerateJWT(u domain.User) (string, error) {
	claims := newJWTClaims(u.Id.String(), u.Role)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, *claims)
	key := GetJWTKey()
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

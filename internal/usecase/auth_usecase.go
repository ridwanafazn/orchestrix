package usecase

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthUsecase interface {
	GenerateGuestToken() (string, error)
}

type authUsecase struct {
	secretKey string
}

func NewAuthUsecase(secret string) AuthUsecase {
	return &authUsecase{secretKey: secret}
}

func (u *authUsecase) GenerateGuestToken() (string, error) {
	// Menggunakan timestamp nano untuk memastikan ID Guest selalu unik
	guestID := fmt.Sprintf("guest-%d", time.Now().UnixNano())

	claims := jwt.MapClaims{
		"user_id": guestID,
		"role":    "guest",
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token berlaku 24 jam
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.secretKey))
}

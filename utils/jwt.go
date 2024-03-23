package utils

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

func GenerateJWTAccessToken(id uuid.UUID, fullname, email, key string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"IdUser":   id.String(),
		"Fullname": fullname,
		"Email":    email,
		"exp":      time.Now().Add(time.Hour * 1).Unix(),
	})
	fmt.Print(token.Claims)
	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return err.Error(), err
	}
	return tokenString, err
}

func GenerateJWTRefreshToken(id uuid.UUID, fullname, email, key string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"IdUser":   id.String(),
		"Fullname": fullname,
		"Email":    email,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return err.Error(), err
	}
	return tokenString, err
}

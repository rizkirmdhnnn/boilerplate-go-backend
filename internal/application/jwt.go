package application

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims.
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func generateJWT(userID int64, email, secret string, exp time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "boilerplate",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}

	return signed, nil
}

// ValidateJWT validates a JWT token and returns the claims.
func ValidateJWT(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse jwt: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid jwt claims")
	}

	return claims, nil
}

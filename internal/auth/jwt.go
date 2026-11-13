package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	tokenEXP = 24 * time.Hour
	issuer   = "shortener"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type ConfigJWT struct {
	SigningKey string
}

type JWTManager struct {
	signingKey []byte
}

func NewJWTManager(cfg ConfigJWT) *JWTManager {
	return &JWTManager{
		signingKey: append([]byte(nil), cfg.SigningKey...),
	}
}

func (m *JWTManager) Generate(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenEXP)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString(m.signingKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (m *JWTManager) Parse(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpercted signing method: %v", t.Header["alg"])
			}

			return m.signingKey, nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("token parse: %v", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	return claims.UserID, nil
}

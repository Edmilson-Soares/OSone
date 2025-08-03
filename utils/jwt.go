package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("secreta")

// Estrutura do token JWT
type Claims struct {
	Id          string   `json:"id"`
	Code        string   `json:"code"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// Função para gerar um token JWT

func VerifyJWT(tokenStr string) (*Claims, bool, error) {

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid {

		return &Claims{}, false, err
	}

	return claims, true, nil

}

func GenerateJWT(id string, code string, role string, permissions []string) (string, error) {
	claims := Claims{
		Id:          id,
		Role:        role,
		Code:        code,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // Expira em 24h
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

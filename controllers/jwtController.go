package controllers

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtController struct {
}

var secret = []byte("my_secret_key")

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (c *JwtController) SignToken() string {
	expirationTime := time.Now().Add(8 * time.Hour)

	claims := &Claims{
		Username: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	// create jwt with payload
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// get signed token
	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		fmt.Printf("failed to sign jwt: %s\n", err)
	}
	return signedToken
}

func (c *JwtController) Validate(signedToken string) bool {
	claims := &Claims{}

	if _, err := jwt.ParseWithClaims(signedToken, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	}); err != nil {
		fmt.Printf("Failed to parse token: %s\n", err)
		return false
	}

	return true
}

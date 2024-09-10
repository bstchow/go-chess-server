package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt"
)

func ParseKeyFunc(verificationKey string) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != "ES256" {
			return nil, fmt.Errorf("unexpected JWT signing method=%v", token.Header["alg"])
		}
		// https://pkg.go.dev/github.com/dgrijalva/jwt-go#ParseECPublicKeyFromPEM
		return jwt.ParseECPublicKeyFromPEM([]byte(verificationKey))
	}
}

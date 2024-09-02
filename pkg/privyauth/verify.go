package privyauth

import (
	"errors"
	"fmt"
	"time"

	"github.com/bstchow/go-chess-server/internal/env"
	"github.com/golang-jwt/jwt"
)

// Defining a Go type for Privy JWTs
type PrivyClaims struct {
	AppId      string `json:"aud,omitempty"`
	Expiration uint64 `json:"exp,omitempty"`
	Issuer     string `json:"iss,omitempty"`
	UserId     string `json:"sub,omitempty"`
}

// This method will be used to check the token's claims later
func (c *PrivyClaims) Valid() error {
	if c.Issuer != "privy.io" {
		return errors.New("iss claim must be 'privy.io'")
	}
	if c.Expiration < uint64(time.Now().Unix()) {
		return errors.New("token is expired")
	}

	return nil
}

func (c *PrivyClaims) ValidateApp(AppId string) error {
	if c.AppId != AppId {
		return errors.New("aud claim must be your Privy App ID")
	}

	return nil
}

func ParseKeyFunc(verificationKey string) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != "ES256" {
			return nil, fmt.Errorf("unexpected JWT signing method=%v", token.Header["alg"])
		}
		// https://pkg.go.dev/github.com/dgrijalva/jwt-go#ParseECPublicKeyFromPEM
		return jwt.ParseECPublicKeyFromPEM([]byte(verificationKey))
	}
}

func AppValidateToken(accessToken string) (privyClaims *PrivyClaims, err error) {
	appId := env.GetEnv("PRIVY_APP_ID")
	verificationKey := env.GetEnv("PRIVY_VERIFICATION_KEY")
	return ValidateToken(accessToken, appId, verificationKey)
}

func ValidateToken(accessToken string, appId string, verificationKey string) (privyClaims *PrivyClaims, err error) {
	// This method will be used to load the verification key in the required format later

	// Check the JWT signature and decode claims
	token, err := jwt.ParseWithClaims(accessToken, &PrivyClaims{}, ParseKeyFunc(verificationKey))
	if err != nil {
		fmt.Println("JWT signature is invalid.")
		return nil, err
	}

	// Parse the JWT claims into your custom struct
	privyClaims, ok := token.Claims.(*PrivyClaims)
	if !ok {
		fmt.Println("JWT does not have all the necessary claims.")
		return nil, errors.New("JWT does not have all the necessary claims")
	}

	// Check the validity of the JWT claims
	err = privyClaims.Valid()
	if err != nil {
		fmt.Printf("JWT claims are invalid, with error=%v.", err)
		return nil, err
	}

	err = privyClaims.ValidateApp(appId)
	if err != nil {
		fmt.Printf("Claimed app is invalid, error=%v.", err)
		return nil, err
	}

	{
		fmt.Println("JWT is valid.")
		fmt.Printf("%v", privyClaims)
	}

	return privyClaims, nil
}

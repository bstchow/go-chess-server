package auth

// TODO: Consolidate validation functions with Privy auth

import (
	"errors"
	"time"

	"github.com/bstchow/go-chess-server/internal/env"
	"github.com/bstchow/go-chess-server/pkg/logging"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

const (
	FC_SIGNER_ADDRESS_USER_ID_SOURCE = "fc_signer_address"
	PRIVY_DID_USER_ID_SOURCE         = "privy_did"
)

// Defining a Go type for JWTs issued by this server
type Claims struct {
	Expiration uint64 `json:"exp,omitempty"`
	Issuer     string `json:"iss,omitempty"`
	UserId     string `json:"sub,omitempty"`
}

// This method will be used to check the token's claims later
func (c *Claims) Valid() error {
	if c.Issuer != env.GetEnv("JWT_ISSUER") {
		return errors.New("iss claim must be" + env.GetEnv("JWT_ISSUER"))
	}
	if c.Expiration < uint64(time.Now().Unix()) {
		return errors.New("token is expired")
	}

	return nil
}

func CreateServerToken(userId string) (tokenString string, err error) {
	expiration := time.Now().Add(time.Hour * 24).Unix()
	claims := &Claims{
		Expiration: uint64(expiration),
		Issuer:     env.GetEnv("JWT_ISSUER"),
		UserId:     userId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	pk, err := jwt.ParseECPrivateKeyFromPEM([]byte((env.GetEnv("SERVER_JWT_PRIVATE_KEY"))))
	if err != nil {
		logging.Error("Encountered error while parsing server private key.")
		return "", err
	}

	signedToken, err := token.SignedString(interface{}(pk))
	if err != nil {
		logging.Error("Encountered error while creating server token.", zap.Error(err))
		return "", err
	}
	return signedToken, nil
}

func ValidateServerTokenDefault(jwtToken string) (claims *Claims, err error) {
	verificationKey := env.GetEnv("SERVER_JWT_PUBLIC_KEY")
	return ValidateServerToken(jwtToken, verificationKey)
}

func ValidateServerToken(jwtToken string, verificationKey string) (claims *Claims, err error) {
	// Check the JWT signature and decode claims
	token, err := jwt.ParseWithClaims(jwtToken, &Claims{}, ParseKeyFunc(verificationKey))
	if err != nil {
		logging.Error("JWT signature is invalid.", zap.Error(err))
		return nil, err
	}

	// Parse the JWT claims into your custom struct
	claims, ok := token.Claims.(*Claims)
	if !ok {
		logging.Error("JWT does not have all the necessary claims.")
		return nil, errors.New("JWT does not have all the necessary claims")
	}

	// Check the validity of the JWT claims
	err = claims.Valid()
	if err != nil {
		logging.Error("JWT claims are invalid, with error=%v.", zap.Error(err))
		return nil, err
	}

	{
		logging.Info("JWT is valid.")
		logging.Info("JWT claims", zap.Any("claims", claims))
	}

	return claims, nil
}

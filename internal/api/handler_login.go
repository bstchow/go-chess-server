package api

import (
	"encoding/json"
	"net/http"

	"github.com/bstchow/go-chess-server/internal/env"
	"github.com/bstchow/go-chess-server/pkg/privyauth"

	"github.com/bstchow/go-chess-server/internal/models"
)

/*
HTTP Handler for when user access login endpoint
*/
func handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		PrivyJWTToken string `json:"privy_jwt_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	var userPrivyDid string
	if env.GetEnv("VALIDATE_PRIVY_JWT") == "true" {
		claims, err := privyauth.AppValidateToken(params.PrivyJWTToken)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid Privy JWT")
			return
		}
		userPrivyDid = claims.UserId
	} else {
		userPrivyDid = params.PrivyJWTToken
	}

	user, err := models.FindOrCreateUser(userPrivyDid)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find user")
		return
	}

	respondWithJSON(w, http.StatusOK, userResponse{
		PlayerPrivyDid: user.PrivyDid,
	})
}

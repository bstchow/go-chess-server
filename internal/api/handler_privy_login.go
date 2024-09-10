package api

import (
	"encoding/json"
	"net/http"

	"github.com/bstchow/go-chess-server/internal/env"
	"github.com/bstchow/go-chess-server/internal/id"
	"github.com/bstchow/go-chess-server/pkg/auth"
)

type PrivyLoginResponse struct {
	UserId   string `json:"user_id"`
	JwtToken string `json:"jwt_token"`
}

/*
HTTP Handler for when user access login endpoint
*/
func handlerPrivyLogin(w http.ResponseWriter, r *http.Request) {
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
		claims, err := auth.PrivyAppValidateToken(params.PrivyJWTToken)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid Privy JWT")
			return
		}
		userPrivyDid = claims.UserId
	} else {
		userPrivyDid = params.PrivyJWTToken
	}

	userId := id.ConstructId(auth.PRIVY_DID_USER_ID_SOURCE, userPrivyDid)
	token, err := auth.CreateServerToken(userId)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT")
		return
	}

	respondWithJSON(w, http.StatusOK, PrivyLoginResponse{
		UserId:   userId,
		JwtToken: token,
	})
}

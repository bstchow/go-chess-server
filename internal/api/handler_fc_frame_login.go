package api

import (
	"encoding/json"
	"net/http"

	"github.com/bstchow/go-chess-server/internal/env"
	"github.com/bstchow/go-chess-server/internal/id"
	"github.com/bstchow/go-chess-server/pkg/auth"
)

type FcFrameLoginResponse struct {
	UserId   string `json:"user_id"`
	JwtToken string `json:"jwt_token"`
}

type FrameRequestClaims struct {
	SigningAddress string `json:"signing_address"`
}

/*
HTTP Handler for when user attempts to authenticate with a signed frame request
*/
func handlerFcFrameLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Message auth.FcFrameMessage `json:"fc_frame_message"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	if env.GetEnv("VALIDATE_FRAME_REQUEST") == "true" {
		ok, err := auth.ValidateFrameMessage(params.Message)

		if err != nil || !ok {
			respondWithError(w, http.StatusUnauthorized, "Invalid Frame Request Signature")
			return
		}
	}

	signingAddress := params.Message.Signer
	id := id.ConstructId(auth.FC_SIGNER_ADDRESS_USER_ID_SOURCE, signingAddress)
	token, err := auth.CreateServerToken(id)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT")
		return
	}

	respondWithJSON(w, http.StatusOK, FcFrameLoginResponse{
		UserId:   id,
		JwtToken: token,
	})
}

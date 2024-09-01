package api

import (
	"net/http"

	"github.com/bstchow/go-chess-server/internal/models"
)

/*
HTTP Handler for when a user wants to get all match records
*/
func handlerSessionGet(w http.ResponseWriter, r *http.Request) {
	player_id := r.URL.Query().Get("player_id")
	if player_id == "" {
		respondWithError(w, http.StatusBadRequest, "Player id not included")
		return
	}

	sessionIDs, err := models.GetSessionsByPlayerID(player_id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid player id")
	}

	respondWithJSON(w, http.StatusOK, sessionIDs)
}

/*
HTTP Handler for when a user wants to get a match record with a specific sesssion id
*/
func handlerSessionGetFromID(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionid")

	session, err := models.GetSessionByID(sessionID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session id")
	}

	respondWithJSON(w, http.StatusOK, session)
}

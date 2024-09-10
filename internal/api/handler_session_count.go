package api

import (
	"net/http"

	"github.com/bstchow/go-chess-server/pkg/agent"
)

type sessionCountResponse struct {
	SessionCount int `json:"session_count"`
}

func injectHandlerSessionCount(agent *agent.Agent) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		respondWithJSON(w, http.StatusOK, sessionCountResponse{
			SessionCount: agent.GetSessionCount(),
		})
	}
}

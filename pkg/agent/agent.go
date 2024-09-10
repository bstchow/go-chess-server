package agent

import (
	"github.com/bstchow/go-chess-server/internal/models"
	"github.com/bstchow/go-chess-server/pkg/auth"
	"github.com/bstchow/go-chess-server/pkg/corenet"
	"github.com/bstchow/go-chess-server/pkg/logging"
	"github.com/bstchow/go-chess-server/pkg/matcher"
	"github.com/bstchow/go-chess-server/pkg/session"
	"github.com/bstchow/go-chess-server/pkg/utils"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Agent struct {
	wsServer *corenet.WebSocketServer
	matcher  *matcher.Matcher
}

// Return an Agent object which is the center module interacting with other modules
func NewAgent() *Agent {
	a := &Agent{
		wsServer: corenet.NewWebSocketServer(),
		matcher:  matcher.NewMatcher(),
	}
	a.wsServer.SetMessageHandler(a.handleWebSocketMessage)
	a.wsServer.SetConnCloseGameHandler(a.playerDisconnectHandler)
	session.SetGameOverHandler(a.handleSessionGameOver)

	return a
}

// Start the server for handling game session
func (a *Agent) StartGameServer() error {
	err := a.wsServer.Start()
	if err != nil {
		return err
	}

	return nil
}

func (a *Agent) GetSessionCount() int {
	return len(a.matcher.SessionMap)
}

/*
Handler for when a game instance ended.
This includes saving the session to the database, close the session
and remove session from tracking of Matcher
*/
func (a *Agent) handleSessionGameOver(s *session.GameSession, sessionID string) {
	players := s.GetPlayers()
	for _, player := range players {
		player.Conn.WriteJSON(struct {
			Type string            `json:"type"`
			Data map[string]string `json:"data"`
		}{
			Type: "endgame",
			Data: map[string]string{
				"game_outcome": s.Game.Outcome().String(),
			},
		})
		player.Conn.Close()
	}
	gameMoves := make([]string, 0, len(s.Game.Moves()))
	for _, move := range s.Game.Moves() {
		gameMoves = append(gameMoves, move.String())
	}
	if _, err := models.InsertSession(sessionID, players[0].ID, players[1].ID, gameMoves); err != nil {
		logging.Error("coulnd't save game", zap.Error(err))
	}
	session.CloseSession(sessionID)
	a.matcher.RemoveSession(players[0].ID, players[1].ID)
}

/*
Handler for when a user connection closes
*/
func (a *Agent) playerDisconnectHandler(connID string) {
	playerId, ok := a.matcher.ConnMap[connID]
	if !ok {
		return
	}

	sessionID, exists := a.matcher.SessionExists(playerId)
	if !exists {
		return
	}

	err := session.PlayerDisconnect(sessionID, playerId)
	if err != nil {
		logging.Warn("player disconnected error",
			zap.String("id", playerId),
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
	}

	delete(a.matcher.ConnMap, connID)

	logging.Info("player disconnected",
		zap.String("id", playerId),
		zap.String("session_id", sessionID),
	)

	// TODO: Auto-resign disconnected player after timeout
}

/*
* Handler for when user socket sends a message
* TODO/SPIKE: Should we add explicit resignation message? Or is that a type of move? Currently, only supported by closing ws connection.
 */
func (a *Agent) handleWebSocketMessage(conn *websocket.Conn, message *corenet.Message, connID *string) {
	type errorResponse struct {
		Type  string `json:"type"`
		Error string `json:"error"`
	}

	jwtToken, ok := message.Data["jwt_token"].(string)
	var playerId string
	claims, authErr := auth.ValidateServerTokenDefault(jwtToken)
	if authErr != nil {
		logging.Info("attempt matchmaking",
			zap.String("status", "rejected"),
			zap.String("error", authErr.Error()),
			zap.String("remote_address", conn.RemoteAddr().String()),
		)
		conn.WriteJSON(errorResponse{
			Type:  "error",
			Error: authErr.Error(),
		})
		return
	}
	playerId = claims.UserId

	switch message.Action {
	case "matching":
		if ok {
			*connID = utils.GenerateUUID()
			logging.Info("attempt matchmaking",
				zap.String("status", "queued"),
				zap.String("id", playerId),
				zap.String("remote_address", conn.RemoteAddr().String()),
			)
			a.matcher.EnterQueue(&session.Player{
				Conn: conn,
				ID:   playerId,
			}, *connID)
		} else {
			logging.Info("attempt matchmaking",
				zap.String("status", "rejected"),
				zap.String("error", "insufficient data"),
				zap.String("remote_address", conn.RemoteAddr().String()),
			)
			conn.WriteJSON(errorResponse{
				Type:  "error",
				Error: "insufficient data",
			})
		}
	case "move":
		sessionID, sessionOK := message.Data["session_id"].(string)
		move, moveOK := message.Data["move"].(string)
		if sessionOK && moveOK {
			logging.Info("attempt making move",
				zap.String("status", "processing"),
				zap.String("id", playerId),
				zap.String("session_id", sessionID),
				zap.String("move", move),
				zap.String("remote_address", conn.RemoteAddr().String()),
			)
			session.ProcessMove(sessionID, playerId, move)
		} else {
			logging.Info("attempt making move",
				zap.String("status", "rejected"),
				zap.String("error", "insufficient data"),
				zap.String("remote_address", conn.RemoteAddr().String()),
			)
			conn.WriteJSON(errorResponse{
				Type:  "error",
				Error: "insufficient data",
			})
		}
	default:
	}
}

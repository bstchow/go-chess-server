package agent

import (
	"github.com/bstchow/go-chess-server/internal/models"
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

/*
Handler for when a game instance ended.
This includes saving the session to the database, close the session
and remove session from tracking of Matcher
*/
func (a *Agent) handleSessionGameOver(s *session.GameSession, sessionID string) {
	playerPrivyDids := make([]string, 0, 2)
	for id, player := range s.Players {
		playerPrivyDids = append(playerPrivyDids, id)
		player.Conn.WriteJSON(struct {
			Type string            `json:"type"`
			Data map[string]string `json:"data"`
		}{
			Type: "endgame",
			Data: map[string]string{
				"game_state": s.Game.GetStatus(),
			},
		})
		player.Conn.Close()
	}
	gameMoves := s.Game.GetAllMoves()
	if _, err := models.InsertSession(sessionID, playerPrivyDids[0], playerPrivyDids[1], gameMoves); err != nil {
		logging.Error("coulnd't save game", zap.Error(err))
	}
	session.CloseSession(sessionID)
	a.matcher.RemoveSession(playerPrivyDids[0], playerPrivyDids[1])
}

/*
Handler for when a user connection closes
*/
func (a *Agent) playerDisconnectHandler(connID string) {
	playerPrivyDid, ok := a.matcher.ConnMap[connID]
	if !ok {
		return
	}

	sessionID, exists := a.matcher.SessionExists(playerPrivyDid)
	if !exists {
		return
	}

	err := session.PlayerLeave(sessionID, playerPrivyDid)
	if err != nil {
		logging.Warn("player disconnected error",
			zap.String("player_privy_did", playerPrivyDid),
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
	}

	delete(a.matcher.ConnMap, connID)

	logging.Info("player disconnected",
		zap.String("player_privy_did", playerPrivyDid),
		zap.String("session_id", sessionID),
	)

	// TODO: Auto-resign disconnected player and persist game session to DB
	//       Need to expose session data from session package
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
	switch message.Action {
	case "matching":
		playerPrivyDid, ok := message.Data["player_privy_did"].(string)
		if ok {
			*connID = utils.GenerateUUID()
			logging.Info("attempt matchmaking",
				zap.String("status", "queued"),
				zap.String("player_privy_did", playerPrivyDid),
				zap.String("remote_address", conn.RemoteAddr().String()),
			)
			a.matcher.EnterQueue(&session.Player{
				Conn: conn,
				ID:   playerPrivyDid,
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
		playerPrivyDid, playerOK := message.Data["player_privy_did"].(string)
		sessionID, sessionOK := message.Data["session_id"].(string)
		move, moveOK := message.Data["move"].(string)
		if playerOK && sessionOK && moveOK {
			logging.Info("attempt making move",
				zap.String("status", "processing"),
				zap.String("player_privy_did", playerPrivyDid),
				zap.String("session_id", sessionID),
				zap.String("move", move),
				zap.String("remote_address", conn.RemoteAddr().String()),
			)
			session.ProcessMove(sessionID, playerPrivyDid, move)
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

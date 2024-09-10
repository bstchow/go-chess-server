package session

import (
	"errors"
	"sync"

	"github.com/bstchow/go-chess-server/pkg/logging"
	"github.com/notnil/chess"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type GameSession struct {
	WhitePlayer *Player
	BlackPlayer *Player
	Game        *chess.Game
}

type SessionResponse struct {
	Type        string      `json:"type"`
	GameState   string      `json:"game_state"`
	PlayerState PlayerState `json:"player_state"`
}

type PlayerState struct {
	IsWhiteSide bool `json:"is_white_side"`
}

var gameSessions = make(map[string]*GameSession)
var mu sync.RWMutex
var gameOverHandler = func(session *GameSession, sessionID string) {
	CloseSession(sessionID)
	session.WhitePlayer.Conn.Close()
	session.BlackPlayer.Conn.Close()
}

func InitSession(sessionID string, whitePlayer *Player, blackPlayer *Player) {
	gameSessions[sessionID] = &GameSession{
		WhitePlayer: whitePlayer,
		BlackPlayer: blackPlayer,
		Game:        chess.NewGame(),
	}
}

func CloseSession(sessionID string) {
	mu.Lock()
	defer mu.Unlock()
	delete(gameSessions, sessionID)
}

func SetGameOverHandler(govHandler func(*GameSession, string)) {
	gameOverHandler = govHandler
}
func StartGame(session *GameSession) {
	for _, player := range []*Player{session.WhitePlayer, session.BlackPlayer} {
		err := player.Conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"start"}`))
		if err != nil {
			logging.Error("Error sending start message", zap.Error(err))
		}
	}
}

func (session *GameSession) GetPlayers() [2]*Player {
	return [2]*Player{session.WhitePlayer, session.BlackPlayer}
}

func (session *GameSession) GetPlayerById(playerID string) (*Player, error) {
	if session.WhitePlayer.ID == playerID {
		return session.WhitePlayer, nil
	} else if session.BlackPlayer.ID == playerID {
		return session.BlackPlayer, nil
	}
	return nil, errors.New("player not in session")
}

func (session *GameSession) GetPlayerSide(playerID string) (bool, error) {
	return session.WhitePlayer.ID == playerID, nil
}

func GetGameFen(sessionID string) (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	session, exists := gameSessions[sessionID]
	if exists {
		return session.Game.FEN(), nil
	}
	return "", errors.New("invalid session id")
}

func GetPlayerState(sessionID, playerID string) (PlayerState, error) {
	mu.RLock()
	defer mu.RUnlock()
	session, exists := gameSessions[sessionID]
	if exists {
		isWhiteSide, err := session.GetPlayerSide(playerID)
		if err != nil {
			return PlayerState{}, errors.New("invalid player id")
		}
		return PlayerState{
			IsWhiteSide: isWhiteSide,
		}, nil
	}
	return PlayerState{}, errors.New("invalid session id")
}

func PlayerInSession(sessionID string, player *Player) bool {
	mu.Lock()
	defer mu.Unlock()
	session, exists := gameSessions[sessionID]
	if exists {
		_, err := session.GetPlayerById(player.ID)
		return err != nil
	}
	return false
}

func PlayerRejoinExisting(sessionID string, player *Player) error {
	mu.Lock()
	defer mu.Unlock()
	session, exists := gameSessions[sessionID]
	if exists {
		if p, err := session.GetPlayerById(player.ID); err == nil {
			if p != nil {
				p.Conn = player.Conn
				logging.Info("Player rejoined session", zap.String("sessionID", sessionID))
				return nil
			}
		}
		return errors.New("player id not in the session")
	}
	return errors.New("invalid session id")
}

func PlayerDisconnect(sessionID, playerID string) error {
	mu.Lock()
	defer mu.Unlock()
	session, exists := gameSessions[sessionID]
	if !exists {
		return errors.New("invalid session id")
	}

	player, err := session.GetPlayerById(playerID)
	if err != nil {
		return err
	}

	player.Conn = nil
	return nil
}

func ProcessMove(sessionID, movingPlayerID, move string) {
	mu.Lock()

	type errorResponse struct {
		Type  string `json:"type"`
		Error string `json:"error"`
	}

	session, exists := gameSessions[sessionID]
	if exists {

		if (session.Game.Position().Turn() == chess.White && session.WhitePlayer.ID != movingPlayerID) ||
			(session.Game.Position().Turn() == chess.Black && session.BlackPlayer.ID != movingPlayerID) {
			logging.Warn("Wrong player moving",
				zap.String("session_id", sessionID),
				zap.String("id", movingPlayerID),
				zap.String("move", move),
			)
			mu.Unlock()
			return
		}

		err := session.Game.MoveStr(move)
		if err != nil {
			logging.Warn("invalid move",
				zap.String("session_id", sessionID),
				zap.String("id", movingPlayerID),
				zap.String("move", move),
				zap.String("error", err.Error()),
			)
			player, err := session.GetPlayerById(movingPlayerID)

			if err := player.Conn.WriteJSON(errorResponse{
				Type:  "error",
				Error: "invalid move: " + err.Error(),
			}); err != nil {
				logging.Info("ws write", zap.Error(err))
			}
			mu.Unlock()
			return
		}

		logging.Info("valid move",
			zap.String("session_id", sessionID),
			zap.String("id", movingPlayerID),
			zap.String("move", move),
		)

		mu.Unlock()

		// notify players about the new board state
		for _, player := range session.GetPlayers() {
			gameFen, err := GetGameFen(sessionID)
			if err != nil {
				logging.Error("invalid session id for game state")
				if err := player.Conn.WriteJSON(errorResponse{
					Type:  "error",
					Error: "coulnd't retrieve game state",
				}); err != nil {
					logging.Info("ws write", zap.Error(err))
				}
				return
			}

			if player == nil {
				continue
			}

			isWhiteSide, err := session.GetPlayerSide(player.ID)
			if err != nil {
				logging.Error("invalid player id")
				continue
			}

			if err := player.Conn.WriteJSON(SessionResponse{
				Type:      "session",
				GameState: gameFen,
				PlayerState: PlayerState{
					IsWhiteSide: isWhiteSide,
				},
			}); err != nil {
				logging.Error("couldn't notify player ", zap.String("id", movingPlayerID))
			}
		}

		if session.Game.Outcome() != chess.NoOutcome {
			gameOverHandler(session, sessionID)
		}
	}
}

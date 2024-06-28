package matcher

import (
	"fmt"
	"sync"
	"time"

	"github.com/yelaco/go-chess-server/internal/session"
	"github.com/yelaco/go-chess-server/pkg/config"
	"github.com/yelaco/go-chess-server/pkg/logging"
	"go.uber.org/zap"
)

type Matcher struct {
	Queue      []*session.Player
	SessionMap map[string]string
	mu         sync.Mutex
}

type matchResponse struct {
	Type        string              `json:"type"`
	SessionID   string              `json:"session_id"`
	GameState   session.GameState   `json:"game_state"`
	PlayerState session.PlayerState `json:"player_state"`
}

type timeoutResponpse struct {
	Type    string `json:"type"`
	Message string `json:"Message"`
}

func NewMatcher() *Matcher {
	return &Matcher{
		Queue:      []*session.Player{},
		SessionMap: map[string]string{},
		mu:         sync.Mutex{},
	}
}

func (m *Matcher) EnterQueue(player *session.Player) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sessionID, exists := m.SessionMap[player.ID]
	if exists {
		m.RejoinMatch(sessionID, player)
		return
	}
	m.Queue = append(m.Queue, player)
	go m.LeaveQueueIfTimeout(player)
	go m.findMatch()
}

func (m *Matcher) LeaveQueueIfTimeout(player *session.Player) {
	time.Sleep(config.MatchingTimeout)
	if player == nil {
		return
	}
	if _, ok := m.SessionMap[player.ID]; !ok {
		player.Conn.WriteJSON(timeoutResponpse{
			Type:    "timeout",
			Message: "Canceled matching due to timeout",
		})
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, p := range m.Queue {
		if p.ID == player.ID || p == player {
			m.Queue = append(m.Queue[:i], m.Queue[i+1:]...)
		}
	}
}

func generateSessionId() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (m *Matcher) findMatch() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Queue) >= 2 {
		player1 := m.Queue[0]
		player2 := m.Queue[1]
		m.Queue = m.Queue[2:]

		sessionID := generateSessionId()
		session.InitSession(sessionID, player1, player2)
		m.SessionMap[player1.ID] = sessionID
		m.SessionMap[player2.ID] = sessionID

		logging.Info("init match",
			zap.String("player_1", player1.ID),
			zap.String("player_2", player2.ID),
		)

		notifyMatchingResult(sessionID, player1)
		notifyMatchingResult(sessionID, player2)
	}
}

func (m *Matcher) RejoinMatch(sessionID string, player *session.Player) {
	notifyMatchingResult(sessionID, player)
}

func notifyMatchingResult(sessionID string, player *session.Player) {
	gameState, err := session.GetGameState(sessionID)
	if err != nil {
		player.Conn.WriteJSON(struct {
			Type  string `json:"type"`
			Error string `json:"error"`
		}{
			Type:  "error",
			Error: "Coulnd't join match: " + err.Error(),
		})
		return
	}

	playerState, err := session.GetPlayerState(sessionID, player.ID)
	if err != nil {
		player.Conn.WriteJSON(struct {
			Type  string `json:"type"`
			Error string `json:"error"`
		}{
			Type:  "error",
			Error: "Coulnd't join match: " + err.Error(),
		})
	}

	session.PlayerJoin(sessionID, player)

	player.Conn.WriteJSON(matchResponse{
		Type:        "matched",
		SessionID:   sessionID,
		GameState:   gameState,
		PlayerState: playerState,
	})
}

func (m *Matcher) SessionExists(playerID string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sessionID, exists := m.SessionMap[playerID]
	if exists {
		return sessionID, true
	} else {
		return "", false
	}
}
func (m *Matcher) RemoveSession(player1, player2 string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.SessionMap, player1)
	delete(m.SessionMap, player2)
}

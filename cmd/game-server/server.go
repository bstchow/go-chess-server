package main

import (
	"github.com/bstchow/go-chess-server/internal/api"
	"github.com/bstchow/go-chess-server/internal/database"
	"github.com/bstchow/go-chess-server/internal/env"
	"github.com/bstchow/go-chess-server/pkg/agent"
	"github.com/bstchow/go-chess-server/pkg/logging"

	"go.uber.org/zap"
)

func main() {
	if !env.ValidateExpectedEnv() {
		logging.Fatal("missing expected environment variables")
	}

	agent := agent.NewAgent()
	database.InitDB()
	defer database.CloseDB()
	// log.Fatal(agent.StartGameServer())

	go func() {
		if err := agent.StartGameServer(); err != nil {
			logging.Fatal("game server failed to start", zap.Error(err))
		}
	}()

	go func() {
		RESTPort := env.GetEnv("REST_PORT")
		if err := api.StartRESTServer(RESTPort); err != nil {
			logging.Fatal("rest server failed to start", zap.Error(err))
		}
	}()

	select {}
}

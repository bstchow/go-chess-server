package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/bstchow/go-chess-server/internal/env"
	"github.com/bstchow/go-chess-server/pkg/logging"

	"go.uber.org/zap"
)

var db *sql.DB

func InitDB() {
	var err error
	user := env.GetEnv("DATABASE_USER")
	password := env.GetEnv("DATABASE_PASSWORD")
	host := env.GetEnv("DATABASE_HOST")
	dbName := env.GetEnv("DATABASE_NAME")
	sslMode := env.GetEnv("SSL_MODE")

	// Assemble the connection string.
	dataSource := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s", host, user, password, dbName, sslMode)

	db, err = sql.Open("postgres", dataSource)
	if err != nil {
		logging.Fatal("database connection failure", zap.Error(err))
	}

	// Ping database to verify the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	logging.Info("database connected")

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
}

func CloseDB() {
	db.Close()
}

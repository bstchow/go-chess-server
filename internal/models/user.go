package models

import (
	"errors"

	"github.com/bstchow/go-chess-server/pkg/utils"
	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"
)

// TODO: Migrate to using GORM for all database interactions.
type User struct {
	gorm.Model
	PlayerID string `json:"player_id" gorm:"unique"`
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`
}

func GetUserByUsername(username string) (User, error) {
	user := User{}
	query := "SELECT player_id, username, password FROM users WHERE username = $1"
	if db == nil {
		return User{}, errors.New("db nil")
	}
	row := db.QueryRow(query, username)
	err := row.Scan(&user.PlayerID, &user.Username, &user.Password)
	return user, err
}

func CreateUser(username, password string) (User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return User{}, err
	}
	user := User{
		PlayerID: utils.GenerateUUID(),
		Username: username,
		Password: string(hashedPassword),
	}

	query := `
        INSERT INTO users (player_id, username, password)
        VALUES ($1, $2, $3)
    `
	_, err = db.Exec(query, user.PlayerID, user.Username, user.Password)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

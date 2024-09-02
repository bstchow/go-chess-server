package models

import (
	"gorm.io/gorm"
)

// TODO: Migrate to using GORM for all database interactions.
type User struct {
	gorm.Model
	PrivyDID string `json:"privy_did" gorm:"uniqueIndex"`
}

func GetUserByPrivyDID(privyDid string) (user User, err error) {
	user = User{}
	result := gormDbWrapper.First(&user, User{PrivyDID: privyDid})
	if err = result.Error; err != nil {
		return user, err
	}

	return user, nil
}

func FindOrCreateUser(privyDid string) (user User, err error) {
	user = User{}
	result := gormDbWrapper.FirstOrCreate(&user, User{PrivyDID: privyDid})
	if err = result.Error; err != nil {
		return user, err
	}

	return user, nil
}

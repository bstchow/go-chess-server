package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Id string `json:"id" gorm:"uniqueIndex"`
}

func GetUserById(id string) (user User, err error) {
	user = User{}
	result := gormDbWrapper.First(&user, User{Id: id})
	if err = result.Error; err != nil {
		return user, err
	}

	return user, nil
}

func FindOrCreateUser(id string) (user User, err error) {
	user = User{}
	result := gormDbWrapper.FirstOrCreate(&user, User{Id: id})
	if err = result.Error; err != nil {
		return user, err
	}

	return user, nil
}

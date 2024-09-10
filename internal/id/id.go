package id

import "fmt"

func ConstructId(userIdSource string, userId string) string {
	return fmt.Sprintf("%s:%s", userIdSource, userId)
}

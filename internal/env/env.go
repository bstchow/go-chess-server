package env

import (
	"log"
	"os"
	"strconv"
)

type EnvValue struct {
	Type         string
	DefaultValue string
}

var EXPECTED_ENV = map[string](EnvValue){
	"WS_PORT":           {"int", "7201"},
	"REST_PORT":         {"int", "7202"},
	"ADMIN_PASSWORD":    {"string", "123"},
	"MATCHING_TIMEOUT":  {"int", "5"},
	"DATABASE_USER":     {"string", "postgres"},
	"DATABASE_PASSWORD": {"string", "postgres"},
	"DATABASE_HOST":     {"string", "localhost"},
	"DATABASE_NAME":     {"string", "postgres"},
	"SSL_MODE":          {"string", "disable"},
}

// GetEnv finds an env variable or the given fallback.
func GetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = EXPECTED_ENV[key].DefaultValue
	}

	return value
}

func ValidateExpectedEnv() bool {
	for key, expectedEnvValue := range EXPECTED_ENV {
		val := GetEnv(key)
		log.Println("Validating key: ", key)

		switch expectedEnvValue.Type {
		case "int":
			_, err := strconv.Atoi(val)
			if err != nil {
				log.Fatal("Expected int for env variable ", key)
			}
		case "string":
			// ENV variables are always strings
			continue
		default:
			log.Fatal("Invalid expected type: . Expect 'int' or 'string'", expectedEnvValue.Type)
		}
	}
	return true
}

func ValidateKey(key string) bool {
	_, exists := os.LookupEnv(key)
	return exists
}

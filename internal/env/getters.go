package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// GetStr retrieves the value of a string env var, else returns a default.
func GetStr(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetInt retrieves the value of an integer env var, else returns a default.
func GetInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strToInt(value)
}

// GetDuration retrieves the value of a Go-parseable duration string env var,
// else returns a default.
func GetDuration(key string, defaultValue string) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return strToDuration(defaultValue)
	}
	return strToDuration(value)
}

// GetBool retrieves the value of a boolean env var, else returns a default.
func GetBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strToBool(value)
}

func strToInt(varStr string) int {
	varInt, err := strconv.Atoi(varStr)
	if err != nil {
		panic(fmt.Errorf("failed to convert var %s to int", varStr))
	}

	return varInt
}

func strToDuration(varStr string) time.Duration {
	duration, err := time.ParseDuration(varStr)
	if err != nil {
		panic(fmt.Errorf("failed to convert var %s to duration", varStr))
	}

	return duration
}

func strToBool(varStr string) bool {
	varBool, err := strconv.ParseBool(varStr)
	if err != nil {
		panic(fmt.Errorf("failed to convert var %s to bool", varStr))
	}

	return varBool
}

package env

import (
	"fmt"
	"os"
	"strconv"
)

func GetString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	fmt.Printf("GetString: key=%s, val=%s, ok=%v\n", key, val, ok) // Debug print

	if !ok {
		return fallback
	}

	return val
}

func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	fmt.Printf("GetInt: key=%s, val=%s, ok=%v\n", key, val, ok) // Debug print

	if !ok {
		return fallback
	}

	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		fmt.Printf("GetInt: error converting val=%s to int: %v\n", val, err) // Debug print
		return fallback
	}

	return valAsInt
}

func GetBool(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	fmt.Printf("GetBool: key=%s, val=%s, ok=%v\n", key, val, ok) // Debug print

	if !ok {
		return fallback
	}

	valAsBool, err := strconv.ParseBool(val)
	if err != nil {
		fmt.Printf("GetBool: error converting val=%s to bool: %v\n", val, err) // Debug print
		return fallback
	}

	return valAsBool
}
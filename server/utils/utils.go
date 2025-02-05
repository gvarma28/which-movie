package utils

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"
)

func ConvertToJSON(inputBytes []byte) (map[string]any, error) {
	var jsonRes map[string]any
	err := json.Unmarshal(inputBytes, &jsonRes)
	if err != nil {
		return nil, err
	}
	return jsonRes, nil
}

func FindInJSON(data any, keys ...string) any {
	current := data
	for _, key := range keys {
		switch currentVal := current.(type) {
		case map[string]any:
			// Handle map traversal
			current = currentVal[key]
		case []any:
			// Handle array traversal
			index, err := strconv.Atoi(key)
			if err != nil || index < 0 || index >= len(currentVal) {
				return nil
			}
			current = currentVal[index]
		default:
			return nil
		}
	}
	return current
}

func WriteToFile(data []byte) {
	err := os.WriteFile("output.txt", data, 0644)
	if err != nil {
		panic(err)
	}
}

// for perfomance testing
func TimeMe(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %v", name, elapsed)
}

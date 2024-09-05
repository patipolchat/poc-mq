package model

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomString(byteLength int) (string, error) {
	// Create a byte slice of the desired length
	bytes := make([]byte, byteLength)

	// Read random bytes into the slice
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Convert the byte slice to a hexadecimal string
	return hex.EncodeToString(bytes), nil
}

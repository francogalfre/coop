package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const randomIDBytes = 16

func randomID() (string, error) {
	raw := make([]byte, randomIDBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("random id: %w", err)
	}

	return hex.EncodeToString(raw), nil
}

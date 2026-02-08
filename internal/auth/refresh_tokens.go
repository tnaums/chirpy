package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func MakeRefreshToken() (string, error) {
	// Note that no error handling is necessary, as Read always succeeds.
	key := make([]byte, 32)
	rand.Read(key)
	// The key can contain any byte value, print the key in hex.
	fmt.Printf("% x\n", key)
	return hex.EncodeToString(key), nil
}

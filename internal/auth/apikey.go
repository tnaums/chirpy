package auth

import (
	"net/http"
	"fmt"
)
func GetAPIKey(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("no token found")
	}
	return token[7:], nil
}

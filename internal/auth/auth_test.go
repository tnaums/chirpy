package auth

import (
	"testing"
	//	"fmt"
)

func TestPassword(t *testing.T) {
	hash, _ := HashPassword("pa$$word")
	got, _ := CheckPasswordHash("pa$$word", hash)
	if got != true {
		t.Errorf("expected 'true' but got %t", got)
	}
}

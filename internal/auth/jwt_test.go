package auth

import (
	"testing"
	"github.com/google/uuid"
	"time"
	"net/http"	
	//	"fmt"
)


func TestMakeJWT(t *testing.T) {
	id := uuid.New()
	sstring := "The perl is in the liver"
	duration := time.Duration(1000000000)

	got, _ := MakeJWT(id, sstring, duration)
	returnid, _ := ValidateJWT(got, sstring)
	if returnid != id {
		t.Errorf("expected %s but got %s", id.String(), returnid.String())
	}
}

func TestValidateJWT(t *testing.T) {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Authorization", "someflykeythatyoucan'tguess")
	headers.Set("Authorization", "someotherthing")
	headers.Set("Authorization", "Bearer isthispositionzero?")	
	got, _ := GetBearerToken(headers)
	if got != "Duh" {
		t.Errorf("expected Duh, but got '%s'", got)
	}
}

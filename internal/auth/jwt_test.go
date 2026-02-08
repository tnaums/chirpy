package auth

import (
	"testing"
	"github.com/google/uuid"
	"time"
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


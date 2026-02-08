package auth

import "testing"

func TestMakeRefreshToken(t *testing.T) {
	t.Run("Asking for a toke", func(t *testing.T) {
		got, _ := MakeRefreshToken()
		want := "myRefreshToken"
		assertCorrectMessage(t, got, want)
	})
}


func assertCorrectMessage(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

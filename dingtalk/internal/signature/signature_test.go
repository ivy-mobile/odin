package signature

import "testing"

func TestSign(t *testing.T) {
	got := Sign(1234567890, "secret")
	want := "EHANl8rkdEjQkqBCtupyHTyMPDmYkU7kJIhqNhHkzV0="
	if got != want {
		t.Fatalf("Sign() = %q, want %q", got, want)
	}
}

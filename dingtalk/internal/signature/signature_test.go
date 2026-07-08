package signature

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSign(t *testing.T) {
	got := Sign(1234567890, "secret")
	require.Equal(t, "EHANl8rkdEjQkqBCtupyHTyMPDmYkU7kJIhqNhHkzV0=", got)
}

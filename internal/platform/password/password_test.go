package password_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/platform/password"
)

func TestHashVerify_RoundTrip(t *testing.T) {
	h, err := password.Hash("correct horse battery staple")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(h, "$argon2id$"))
	require.NoError(t, password.Verify("correct horse battery staple", h))
}

func TestVerify_WrongPassword(t *testing.T) {
	h, err := password.Hash("the right one")
	require.NoError(t, err)
	require.ErrorIs(t, password.Verify("the wrong one", h), password.ErrMismatch)
}

func TestHash_SaltIsRandom(t *testing.T) {
	a, _ := password.Hash("same")
	b, _ := password.Hash("same")
	require.NotEqual(t, a, b)
}

func TestVerify_MalformedHash(t *testing.T) {
	for _, bad := range []string{"", "plaintext", "$argon2id$v=19$bad", "$scrypt$x$y$z$w$v"} {
		err := password.Verify("x", bad)
		require.Error(t, err)
		require.NotErrorIs(t, err, password.ErrMismatch)
	}
}

// Package password hashes and verifies account passwords with Argon2id, using the
// PHC string format so the parameters travel with the hash (ADR-0004). Parameters
// are the planning.md build-time defaults, pending the M1 security review.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Params are the Argon2id cost parameters.
type Params struct {
	Memory      uint32 // KiB
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// Default is the current parameter set (19 MiB, t=2, p=1). Revisit at the M1
// security review on the target host.
var Default = Params{
	Memory:      19 * 1024,
	Iterations:  2,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}

// ErrMismatch is returned when a password does not match the hash.
var ErrMismatch = errors.New("password does not match")

// Hash returns a PHC-format Argon2id hash of plain using the default parameters.
func Hash(plain string) (string, error) { return hashWith(plain, Default) }

func hashWith(plain string, p Params) (string, error) {
	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	key := argon2.IDKey([]byte(plain), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Iterations, p.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// Verify reports whether plain matches the given PHC-format hash. It returns
// ErrMismatch on a valid-but-wrong password and a different error if the hash is
// malformed.
func Verify(plain, encoded string) error {
	p, salt, key, err := decode(encoded)
	if err != nil {
		return err
	}
	//nolint:gosec // G115: key length comes from a decoded 32-byte hash, well within uint32
	other := argon2.IDKey([]byte(plain), salt, p.Iterations, p.Memory, p.Parallelism, uint32(len(key)))
	if subtle.ConstantTimeCompare(key, other) == 1 {
		return nil
	}
	return ErrMismatch
}

func decode(encoded string) (Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Params{}, nil, nil, errors.New("password: not an argon2id hash")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return Params{}, nil, nil, errors.New("password: unsupported argon2 version")
	}

	var p Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism); err != nil {
		return Params{}, nil, nil, errors.New("password: malformed parameters")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Params{}, nil, nil, errors.New("password: malformed salt")
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Params{}, nil, nil, errors.New("password: malformed key")
	}
	return p, salt, key, nil
}

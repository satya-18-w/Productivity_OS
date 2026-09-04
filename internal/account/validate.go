package account

import (
	"net/mail"
	"strings"

	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

// Password policy (Q6 default): length only, 12–128 bytes.
const (
	minPasswordLen = 12
	maxPasswordLen = 128
	maxEmailLen    = 254
)

// DefaultTimezone is applied when registration supplies no timezone (Q4).
const DefaultTimezone = "UTC"

func validateEmail(raw string) (string, string) {
	e := strings.TrimSpace(raw)
	if e == "" {
		return "", "email is required"
	}
	if len(e) > maxEmailLen {
		return "", "email is too long"
	}
	addr, err := mail.ParseAddress(e)
	if err != nil || addr.Name != "" {
		return "", "email is not a valid address"
	}
	return strings.ToLower(addr.Address), ""
}

func validatePassword(p string) string {
	switch {
	case len(p) < minPasswordLen:
		return "password must be at least 12 characters"
	case len(p) > maxPasswordLen:
		return "password must be at most 128 characters"
	default:
		return ""
	}
}

// resolveTimezone applies the Q4 rule: empty -> default; present-but-invalid -> error.
func resolveTimezone(tz string) (string, string) {
	if strings.TrimSpace(tz) == "" {
		return DefaultTimezone, ""
	}
	if !timezone.Valid(tz) {
		return "", "timezone is not a valid IANA name"
	}
	return tz, ""
}

func validateRegister(in RegisterInput) (RegisterInput, *ValidationError) {
	fields := map[string]string{}

	email, msg := validateEmail(in.Email)
	if msg != "" {
		fields["email"] = msg
	}
	if msg := validatePassword(in.Password); msg != "" {
		fields["password"] = msg
	}
	tz, msg := resolveTimezone(in.Timezone)
	if msg != "" {
		fields["timezone"] = msg
	}

	if len(fields) > 0 {
		return RegisterInput{}, &ValidationError{Fields: fields}
	}
	return RegisterInput{Email: email, Password: in.Password, Timezone: tz}, nil
}

package account_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/satya-18-w/productivity-os/internal/account"
	"github.com/satya-18-w/productivity-os/internal/platform/pgtest"
)

func newSvc(t *testing.T, ttl time.Duration) (account.Service, *pgxpool.Pool) {
	t.Helper()
	pool := pgtest.Pool(t)
	return account.NewService(pool, ttl), pool
}

func mustRegister(t *testing.T, svc account.Service, email, pw, tz string) account.Session {
	t.Helper()
	_, sess, err := svc.Register(context.Background(), account.RegisterInput{
		Email: email, Password: pw, Timezone: tz,
	})
	require.NoError(t, err)
	return sess
}

func TestRegister_CreatesAccountAndSession(t *testing.T) {
	svc, pool := newSvc(t, time.Hour)
	ctx := context.Background()

	profile, sess, err := svc.Register(ctx, account.RegisterInput{
		Email: "Alice@Example.com", Password: "correct horse staple", Timezone: "Asia/Kolkata",
	})
	require.NoError(t, err)
	require.Equal(t, "alice@example.com", profile.Email)
	require.Equal(t, "Asia/Kolkata", profile.Timezone)
	require.NotEmpty(t, sess.Token)
	require.True(t, sess.ExpiresAt.After(time.Now()))

	var count int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM sessions s JOIN accounts a ON a.id = s.account_id
		 WHERE a.email = $1 AND s.expires_at > now()`, profile.Email).Scan(&count))
	require.Equal(t, 1, count)

	// The raw token is never stored verbatim.
	var stored int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM sessions WHERE token = $1`, sess.Token).Scan(&stored))
	require.Equal(t, 0, stored, "session table stores only the token hash")
}

func TestRegister_StoresOnlyArgon2idHash(t *testing.T) {
	svc, pool := newSvc(t, time.Hour)
	const pw = "super secret value 123"
	mustRegister(t, svc, "bob@example.com", pw, "UTC")

	var stored string
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT password_hash FROM accounts WHERE email = 'bob@example.com'`).Scan(&stored))
	require.True(t, strings.HasPrefix(stored, "$argon2id$"))
	require.NotContains(t, stored, pw)
}

func TestRegister_DuplicateEmailCaseInsensitive(t *testing.T) {
	svc, pool := newSvc(t, time.Hour)
	ctx := context.Background()
	mustRegister(t, svc, "dup@example.com", "password one two", "UTC")

	_, _, err := svc.Register(ctx, account.RegisterInput{
		Email: "DUP@example.com", Password: "another password", Timezone: "UTC",
	})
	require.ErrorIs(t, err, account.ErrEmailTaken)

	var n int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM accounts WHERE email = 'dup@example.com'`).Scan(&n))
	require.Equal(t, 1, n)
}

func TestRegister_ValidationErrors(t *testing.T) {
	svc, _ := newSvc(t, time.Hour)
	cases := []struct {
		name  string
		in    account.RegisterInput
		field string
	}{
		{"bad email", account.RegisterInput{Email: "nope", Password: "long enough pw", Timezone: "UTC"}, "email"},
		{"short password", account.RegisterInput{Email: "a@b.com", Password: "short", Timezone: "UTC"}, "password"},
		{"bad timezone", account.RegisterInput{Email: "a@b.com", Password: "long enough pw", Timezone: "Mars/Base"}, "timezone"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := svc.Register(context.Background(), tc.in)
			var verr *account.ValidationError
			require.ErrorAs(t, err, &verr)
			require.Contains(t, verr.Fields, tc.field)
		})
	}
}

func TestRegister_DefaultTimezoneWhenAbsent(t *testing.T) {
	svc, _ := newSvc(t, time.Hour)
	profile, _, err := svc.Register(context.Background(), account.RegisterInput{
		Email: "tz@example.com", Password: "a decent password", Timezone: "",
	})
	require.NoError(t, err)
	require.Equal(t, "UTC", profile.Timezone)
}

func TestAuthenticate(t *testing.T) {
	svc, _ := newSvc(t, time.Hour)
	ctx := context.Background()
	mustRegister(t, svc, "auth@example.com", "the real password", "UTC")

	sess, err := svc.Authenticate(ctx, "AUTH@example.com", "the real password")
	require.NoError(t, err)
	require.NotEmpty(t, sess.Token)

	_, err = svc.Authenticate(ctx, "auth@example.com", "wrong password here")
	require.ErrorIs(t, err, account.ErrInvalidCredentials)

	_, err = svc.Authenticate(ctx, "ghost@example.com", "the real password")
	require.ErrorIs(t, err, account.ErrInvalidCredentials)
}

func TestResolveSession(t *testing.T) {
	svc, _ := newSvc(t, time.Hour)
	ctx := context.Background()
	_, sess, err := svc.Register(ctx, account.RegisterInput{
		Email: "res@example.com", Password: "resolve me please", Timezone: "UTC",
	})
	require.NoError(t, err)

	id, err := svc.ResolveSession(ctx, sess.Token)
	require.NoError(t, err)
	require.NotEqual(t, "00000000-0000-0000-0000-000000000000", id.AccountID.String())

	_, err = svc.ResolveSession(ctx, "not-a-real-token")
	require.ErrorIs(t, err, account.ErrSessionInvalid)
}

func TestResolveSession_Expired(t *testing.T) {
	svc, _ := newSvc(t, 30*time.Millisecond)
	ctx := context.Background()
	_, sess, err := svc.Register(ctx, account.RegisterInput{
		Email: "exp@example.com", Password: "expire me soon ok", Timezone: "UTC",
	})
	require.NoError(t, err)

	time.Sleep(60 * time.Millisecond)
	_, err = svc.ResolveSession(ctx, sess.Token)
	require.ErrorIs(t, err, account.ErrSessionInvalid)
}

func TestEndSession(t *testing.T) {
	svc, _ := newSvc(t, time.Hour)
	ctx := context.Background()
	_, sess, _ := svc.Register(ctx, account.RegisterInput{
		Email: "logout@example.com", Password: "log me out now ok", Timezone: "UTC",
	})

	require.NoError(t, svc.EndSession(ctx, sess.Token))
	_, err := svc.ResolveSession(ctx, sess.Token)
	require.ErrorIs(t, err, account.ErrSessionInvalid)
}

func TestChangePassword_EndsAllSessions(t *testing.T) {
	svc, pool := newSvc(t, time.Hour)
	ctx := context.Background()
	_, sess, _ := svc.Register(ctx, account.RegisterInput{
		Email: "cp@example.com", Password: "original password!", Timezone: "UTC",
	})
	id, err := svc.ResolveSession(ctx, sess.Token)
	require.NoError(t, err)

	require.ErrorIs(t,
		svc.ChangePassword(ctx, id.AccountID, "wrong current pw", "brand new password"),
		account.ErrInvalidCredentials)

	require.NoError(t, svc.ChangePassword(ctx, id.AccountID, "original password!", "brand new password"))

	var open int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM sessions WHERE account_id = $1`, id.AccountID).Scan(&open))
	require.Equal(t, 0, open)

	_, err = svc.Authenticate(ctx, "cp@example.com", "brand new password")
	require.NoError(t, err)
}

func TestSetTimezone(t *testing.T) {
	svc, _ := newSvc(t, time.Hour)
	ctx := context.Background()
	_, sess, _ := svc.Register(ctx, account.RegisterInput{
		Email: "stz@example.com", Password: "timezone changer!", Timezone: "UTC",
	})
	id, _ := svc.ResolveSession(ctx, sess.Token)

	require.NoError(t, svc.SetTimezone(ctx, id.AccountID, "America/New_York"))
	p, err := svc.Read(ctx, id.AccountID)
	require.NoError(t, err)
	require.Equal(t, "America/New_York", p.Timezone)

	var verr *account.ValidationError
	require.ErrorAs(t, svc.SetTimezone(ctx, id.AccountID, "Mars/Base"), &verr)
	require.ErrorAs(t, svc.SetTimezone(ctx, id.AccountID, ""), &verr)
}

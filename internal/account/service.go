package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/satya-18-w/productivity-os/internal/account/accountdb"
	"github.com/satya-18-w/productivity-os/internal/platform/password"
	"github.com/satya-18-w/productivity-os/internal/platform/timezone"
)

const uniqueViolation = "23505"

type service struct {
	pool       *pgxpool.Pool
	q          *accountdb.Queries
	sessionTTL time.Duration
	now        func() time.Time
}

// NewService builds the account service over a connection pool.
func NewService(pool *pgxpool.Pool, sessionTTL time.Duration) Service {
	return &service{
		pool:       pool,
		q:          accountdb.New(pool),
		sessionTTL: sessionTTL,
		now:        time.Now,
	}
}

func (s *service) Register(ctx context.Context, raw RegisterInput) (Profile, Session, error) {
	in, verr := validateRegister(raw)
	if verr != nil {
		return Profile{}, Session{}, verr
	}

	hash, err := password.Hash(in.Password)
	if err != nil {
		return Profile{}, Session{}, fmt.Errorf("hash password: %w", err)
	}

	var (
		profile Profile
		sess    Session
	)
	err = s.inTx(ctx, func(q *accountdb.Queries) error {
		acc, err := q.CreateAccount(ctx, accountdb.CreateAccountParams{
			Email:        in.Email,
			PasswordHash: hash,
			Timezone:     in.Timezone,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
				return ErrEmailTaken
			}
			return fmt.Errorf("create account: %w", err)
		}
		profile = Profile{Email: acc.Email, Timezone: acc.Timezone}
		sess, err = s.openSession(ctx, q, acc.ID)
		return err
	})
	if err != nil {
		return Profile{}, Session{}, err
	}
	return profile, sess, nil
}

func (s *service) Authenticate(ctx context.Context, email, plaintext string) (Session, error) {
	normalized, msg := validateEmail(email)
	if msg != "" {
		return Session{}, ErrInvalidCredentials
	}

	acc, err := s.q.GetAccountByEmail(ctx, normalized)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Still spend time verifying to blunt timing differences.
			_ = password.Verify(plaintext, decoyHash)
			return Session{}, ErrInvalidCredentials
		}
		return Session{}, fmt.Errorf("lookup account: %w", err)
	}

	if err := password.Verify(plaintext, acc.PasswordHash); err != nil {
		return Session{}, ErrInvalidCredentials
	}

	var sess Session
	err = s.inTx(ctx, func(q *accountdb.Queries) error {
		var e error
		sess, e = s.openSession(ctx, q, acc.ID)
		return e
	})
	if err != nil {
		return Session{}, err
	}
	return sess, nil
}

func (s *service) ResolveSession(ctx context.Context, token string) (Identity, error) {
	if token == "" {
		return Identity{}, ErrSessionInvalid
	}
	row, err := s.q.GetSession(ctx, hashToken(token))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Identity{}, ErrSessionInvalid
		}
		return Identity{}, fmt.Errorf("get session: %w", err)
	}
	if !row.ExpiresAt.Valid || !s.now().Before(row.ExpiresAt.Time) {
		return Identity{}, ErrSessionInvalid
	}
	// Best-effort: a failed last-seen update must not fail the request.
	if err := s.q.TouchSession(ctx, hashToken(token)); err != nil {
		slog.WarnContext(ctx, "touch session failed", slog.String("error", err.Error()))
	}
	return Identity{AccountID: row.AccountID}, nil
}

func (s *service) EndSession(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := s.q.DeleteSession(ctx, hashToken(token)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *service) EndAllSessions(ctx context.Context, accountID uuid.UUID) error {
	if err := s.q.DeleteAccountSessions(ctx, accountID); err != nil {
		return fmt.Errorf("delete account sessions: %w", err)
	}
	return nil
}

func (s *service) ChangePassword(ctx context.Context, accountID uuid.UUID, current, next string) error {
	currentHash, err := s.q.GetAccountPasswordHash(ctx, accountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAccountNotFound
		}
		return fmt.Errorf("get password hash: %w", err)
	}
	if err := password.Verify(current, currentHash); err != nil {
		return ErrInvalidCredentials
	}
	if msg := validatePassword(next); msg != "" {
		return &ValidationError{Fields: map[string]string{"new_password": msg}}
	}

	newHash, err := password.Hash(next)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.inTx(ctx, func(q *accountdb.Queries) error {
		if err := q.UpdateAccountPassword(ctx, accountdb.UpdateAccountPasswordParams{
			ID: accountID, PasswordHash: newHash,
		}); err != nil {
			return fmt.Errorf("update password: %w", err)
		}
		if err := q.DeleteAccountSessions(ctx, accountID); err != nil {
			return fmt.Errorf("end sessions: %w", err)
		}
		return nil
	})
}

func (s *service) SetTimezone(ctx context.Context, accountID uuid.UUID, tz string) error {
	// An explicit set requires a real IANA name; "" is not "use the default" here.
	if strings.TrimSpace(tz) == "" {
		return &ValidationError{Fields: map[string]string{"timezone": "timezone is required"}}
	}
	if !timezone.Valid(tz) {
		return &ValidationError{Fields: map[string]string{"timezone": "timezone is not a valid IANA name"}}
	}
	if err := s.q.UpdateAccountTimezone(ctx, accountdb.UpdateAccountTimezoneParams{
		ID: accountID, Timezone: tz,
	}); err != nil {
		return fmt.Errorf("update timezone: %w", err)
	}
	return nil
}

func (s *service) Read(ctx context.Context, accountID uuid.UUID) (Profile, error) {
	row, err := s.q.GetAccountByID(ctx, accountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, ErrAccountNotFound
		}
		return Profile{}, fmt.Errorf("get account: %w", err)
	}
	return Profile{Email: row.Email, Timezone: row.Timezone}, nil
}

// openSession creates a session row and returns the issued token. It must be
// called inside a transaction q.
func (s *service) openSession(ctx context.Context, q *accountdb.Queries, accountID uuid.UUID) (Session, error) {
	token, err := newToken()
	if err != nil {
		return Session{}, err
	}
	expires := s.now().Add(s.sessionTTL)
	// Only the hash of the token is stored; a database dump does not yield usable
	// session tokens.
	if _, err := q.CreateSession(ctx, accountdb.CreateSessionParams{
		Token:     hashToken(token),
		AccountID: accountID,
		ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true},
	}); err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	return Session{Token: token, ExpiresAt: expires}, nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *service) inTx(ctx context.Context, fn func(*accountdb.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(s.q.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// decoyHash is a valid Argon2id hash of a random value, used to keep the
// unknown-email path roughly as expensive as the wrong-password path.
var decoyHash = mustHash()

func mustHash() string {
	h, err := password.Hash("decoy-" + time.Now().String())
	if err != nil {
		panic(err)
	}
	return h
}

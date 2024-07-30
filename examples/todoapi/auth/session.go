package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"fx.prodigy9.co/data"
)

const SessionTokenBytes = 32
const DefaultSessionAge = 7 * 24 * time.Hour

type Session struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func GenerateSessionToken() (string, error) {
	var buf [SessionTokenBytes]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	} else {
		return base64.URLEncoding.EncodeToString(buf[:]), nil
	}
}

func GetSessionByToken(ctx context.Context, token string) (sess *Session, err error) {
	scope, cancel, err := data.NewScopeErr(ctx, &err)
	defer cancel()
	if err != nil {
		return nil, err
	}

	sess, sql := &Session{}, `SELECT * FROM sessions WHERE token = $1 LIMIT 1`
	if err = scope.Get(sess, sql, token); err != nil {
		return nil, err
	} else {
		return sess, nil
	}
}

func (s *Session) IsExpired() bool {
	return !s.ExpiresAt.IsZero() &&
		s.ExpiresAt.Before(time.Now())
}

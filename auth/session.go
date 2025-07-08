package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type Session struct {
	ID         string `db:"id"`
	OAuthState string `db:"oauth_state"`
	UserID     string `db:"user_id"`
	Avatar     string `db:"avatar"`
}

type sessionCtxKeyT string

var sessionCtxKey = sessionCtxKeyT("session")

func contextWithSession(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey, session)
}

func SessionFromContext(ctx context.Context) (Session, bool) {
	v := ctx.Value(sessionCtxKey)
	if v == nil {
		return Session{}, false
	}
	session, ok := v.(Session)
	return session, ok
}

func MustSessionFromContext(ctx context.Context) Session {
	s, ok := SessionFromContext(ctx)
	if !ok {
		panic("missing or invalid session")
	}
	return s
}

func generateRandomSession() (session Session, err error) {
	session.ID, err = generateRandomString()
	if err != nil {
		return
	}
	session.OAuthState, err = generateRandomString()
	return
}

func generateRandomString() (string, error) {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

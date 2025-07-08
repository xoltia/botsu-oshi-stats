package auth

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type SessionStore struct {
	db *sqlx.DB
}

func CreateSessionStore(ctx context.Context, db *sqlx.DB) (*SessionStore, error) {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS sessions (
			id          TEXT NOT NULL PRIMARY KEY,
			oauth_state TEXT NOT NULL,
			user_id     TEXT NOT NULL,
			avatar      TEXT NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}
	return &SessionStore{db}, nil
}

func (s *SessionStore) setSessionUserData(ctx context.Context, sessionID string, userID string, avatar string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE sessions
		SET user_id = ?,
			avatar = ?
		WHERE id = ?
	`, userID, avatar, sessionID)
	return err
}

func (s *SessionStore) insertSession(ctx context.Context, session Session) error {
	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO sessions (id, oauth_state, user_id, avatar)
		VALUES (:id, :oauth_state, :user_id, :avatar)
	`, session)
	return err
}

func (s *SessionStore) findSession(ctx context.Context, sid string) (Session, error) {
	var session Session
	err := s.db.GetContext(ctx, &session, "SELECT * FROM sessions WHERE id = ?", sid)
	if err != nil {
		return Session{}, err
	}
	return session, nil
}

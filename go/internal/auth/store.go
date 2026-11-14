package auth

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"time"
)

type SQLSessionStore struct{ DB *sql.DB }

func (s SQLSessionStore) CreateSession(ctx context.Context, v Session) error {
	_, e := s.DB.ExecContext(ctx, `INSERT INTO "SantaGoSessions" ("Id","UserId","TokenHash","CsrfHash","ExpiresAt","CreatedAt") VALUES ($1,$2,$3,$4,$5,NOW())`, v.ID, v.UserID, v.TokenHash, v.CSRFHash, v.ExpiresAt)
	return e
}
func (s SQLSessionStore) GetSession(ctx context.Context, h string) (Session, error) {
	var v Session
	e := s.DB.QueryRowContext(ctx, `SELECT "Id","UserId","TokenHash","CsrfHash","ExpiresAt" FROM "SantaGoSessions" WHERE "TokenHash"=$1`, h).Scan(&v.ID, &v.UserID, &v.TokenHash, &v.CSRFHash, &v.ExpiresAt)
	return v, e
}
func (s SQLSessionStore) DeleteSession(ctx context.Context, h string) error {
	_, e := s.DB.ExecContext(ctx, `DELETE FROM "SantaGoSessions" WHERE "TokenHash"=$1`, h)
	return e
}

var ErrInvalidUser = errors.New("invalid user")

type User struct {
	ID                                                                              uuid.UUID
	Email, NormalizedEmail, UserName, NormalizedUserName, PasswordHash, DisplayName string
	AccessFailedCount                                                               int
	LockoutEnd                                                                      *time.Time
	LockoutEnabled                                                                  bool
}

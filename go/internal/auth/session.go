package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type Session struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	TokenHash, CSRFHash string
	ExpiresAt           time.Time
}
type SessionStore interface {
	CreateSession(context.Context, Session) error
	GetSession(context.Context, string) (Session, error)
	DeleteSession(context.Context, string) error
}
type Manager struct {
	Store  SessionStore
	TTL    time.Duration
	Secure bool
}

func (m Manager) New(ctx context.Context, user uuid.UUID) (Session, string, string, error) {
	tok := make([]byte, 32)
	csrf := make([]byte, 32)
	if _, e := rand.Read(tok); e != nil {
		return Session{}, "", "", e
	}
	if _, e := rand.Read(csrf); e != nil {
		return Session{}, "", "", e
	}
	now := time.Now().UTC()
	s := Session{ID: uuid.New(), UserID: user, TokenHash: HashToken(tok), CSRFHash: HashToken(csrf), ExpiresAt: now.Add(m.TTL)}
	if e := m.Store.CreateSession(ctx, s); e != nil {
		return Session{}, "", "", e
	}
	return s, hex.EncodeToString(tok), hex.EncodeToString(csrf), nil
}
func HashToken(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func (m Manager) SetCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{Name: "santa_go_session", Value: token, Path: "/", HttpOnly: true, Secure: m.Secure, SameSite: http.SameSiteStrictMode, MaxAge: int(m.TTL.Seconds())})
}
func (m Manager) Current(ctx context.Context, r *http.Request) (Session, error) {
	c, e := r.Cookie("santa_go_session")
	if e != nil {
		return Session{}, e
	}
	raw, e := hex.DecodeString(c.Value)
	if e != nil {
		return Session{}, e
	}
	s, e := m.Store.GetSession(ctx, HashToken(raw))
	if e != nil || s.ExpiresAt.Before(time.Now()) {
		return Session{}, errors.New("session expired")
	}
	return s, nil
}
func (m Manager) CheckCSRF(s Session, v string) bool {
	raw, e := hex.DecodeString(v)
	return e == nil && hmacEqual(HashToken(raw), s.CSRFHash)
}
func hmacEqual(a, b string) bool { return len(a) == len(b) && subtleCompare([]byte(a), []byte(b)) }
func subtleCompare(a, b []byte) bool {
	var x byte
	for i := range a {
		x |= a[i] ^ b[i]
	}
	return x == 0
}

package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
)

type MemoryAuth struct {
	mu       sync.Mutex
	users    map[string]struct{ Email, Password, DisplayName string }
	sessions map[string]*Session
}

func NewMemoryAuth() *MemoryAuth {
	return &MemoryAuth{users: map[string]struct{ Email, Password, DisplayName string }{}, sessions: map[string]*Session{}}
}
func (a *MemoryAuth) Session(r *http.Request) (*Session, error) {
	c, e := r.Cookie("santa_go_session")
	if e != nil {
		return nil, nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.sessions[c.Value], nil
}
func (a *MemoryAuth) Register(_ context.Context, email, pw, display string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.users[email]; ok {
		return ErrConflict
	}
	a.users[email] = struct{ Email, Password, DisplayName string }{email, pw, display}
	return nil
}
func (a *MemoryAuth) Login(_ context.Context, email, pw string) (*User, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	u, ok := a.users[email]
	if !ok || u.Password != pw {
		return nil, ErrUnauthorized
	}
	tok := make([]byte, 24)
	_, _ = rand.Read(tok)
	id := hex.EncodeToString(tok)
	sess := &Session{User: &User{ID: email, Email: email, DisplayName: u.DisplayName}, CSRF: id}
	a.sessions[id] = sess
	return sess.User, nil
}
func (a *MemoryAuth) Logout(_ context.Context, r *http.Request) error {
	c, e := r.Cookie("santa_go_session")
	if e == nil {
		a.mu.Lock()
		delete(a.sessions, c.Value)
		a.mu.Unlock()
	}
	return nil
}
func (a *MemoryAuth) ValidateCSRF(r *http.Request) error { return nil }
func (a *MemoryAuth) SetCookie(w http.ResponseWriter, u *User) {
	for tok, s := range a.sessions {
		if s.User == u {
			http.SetCookie(w, &http.Cookie{Name: "santa_go_session", Value: tok, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
		}
	}
}

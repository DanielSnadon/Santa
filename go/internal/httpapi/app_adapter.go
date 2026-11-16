package httpapi

import (
	"context"
	"github.com/google/uuid"
	"santa/go/internal/app"
	"strings"
)

// AppAdapter exposes the domain service through the HTTP port's string IDs.
type AppAdapter struct{ Service app.Service }

func appError(err error) error {
	if err == nil {
		return nil
	}
	m := err.Error()
	if strings.Contains(m, "not found") {
		return ErrNotFound
	}
	if strings.Contains(m, "conflict") {
		return ErrConflict
	}
	if strings.Contains(m, "invalid") || strings.Contains(m, "unknown account") {
		return ErrValidation
	}
	if strings.Contains(m, "unsatisfiable") || strings.Contains(m, "search_limit") || strings.Contains(m, "resource_limit") {
		return ErrValidation
	}
	return err
}
func (a AppAdapter) actor(u *User) (uuid.UUID, error) {
	if u == nil {
		return uuid.Nil, ErrUnauthorized
	}
	id, e := uuid.Parse(u.ID)
	if e != nil {
		return uuid.Nil, ErrUnauthorized
	}
	return id, nil
}
func (a AppAdapter) List(ctx context.Context, u *User) ([]Game, error) {
	id, e := a.actor(u)
	if e != nil {
		return nil, e
	}
	v, e := a.Service.List(ctx, id)
	if e != nil {
		return nil, appError(e)
	}
	out := make([]Game, 0, len(v))
	for _, g := range v {
		out = append(out, Game{ID: g.ID.String(), Title: g.Title, State: string(g.State), Version: g.Version, Owner: g.IsOwner})
	}
	return out, nil
}
func (a AppAdapter) Get(ctx context.Context, u *User, id string) (Game, error) {
	actor, e := a.actor(u)
	gid, e2 := uuid.Parse(id)
	if e != nil {
		return Game{}, e
	}
	if e2 != nil {
		return Game{}, ErrNotFound
	}
	g, e := a.Service.Get(ctx, actor, gid)
	if e != nil {
		return Game{}, appError(e)
	}
	return Game{ID: g.ID.String(), Title: g.Title, State: string(g.State), Version: g.Version, Owner: g.IsOwner}, nil
}
func (a AppAdapter) Configuration(ctx context.Context, u *User, id string) (Configuration, error) {
	actor, e := a.actor(u)
	gid, e2 := uuid.Parse(id)
	if e != nil {
		return Configuration{}, e
	}
	if e2 != nil {
		return Configuration{}, ErrNotFound
	}
	c, e := a.Service.Configuration(ctx, actor, gid)
	if e != nil {
		return Configuration{}, appError(e)
	}
	out := Configuration{}
	for _, p := range c.Participants {
		out.Participants = append(out.Participants, Participant{ID: p.ID.String(), Email: p.Email, DisplayName: p.Name})
	}
	for _, p := range c.Exclusions {
		out.Exclusions = append(out.Exclusions, Pair{First: p.First, Second: p.Second})
	}
	return out, nil
}
func (a AppAdapter) Create(ctx context.Context, u *User, title string) (string, error) {
	actor, e := a.actor(u)
	if e != nil {
		return "", e
	}
	id, e := a.Service.Create(ctx, actor, title)
	return id.String(), e
}
func (a AppAdapter) Configure(ctx context.Context, u *User, id string, c Configuration, ver int64) error {
	actor, e := a.actor(u)
	gid, e2 := uuid.Parse(id)
	if e != nil {
		return e
	}
	if e2 != nil {
		return ErrNotFound
	}
	in := app.ConfigurationInput{Version: ver}
	for _, p := range c.Participants {
		in.Emails = append(in.Emails, p.Email)
	}
	if len(in.Emails) == 0 {
		in.Emails = append(in.Emails, c.Emails...)
	}
	for _, p := range c.Exclusions {
		in.Exclusions = append(in.Exclusions, app.Exclusion{First: p.First, Second: p.Second})
	}
	return appError(a.Service.Configure(ctx, actor, gid, in))
}
func (a AppAdapter) Publish(ctx context.Context, u *User, id string, in PublishRequest) error {
	actor, e := a.actor(u)
	gid, e2 := uuid.Parse(id)
	if e != nil {
		return e
	}
	if e2 != nil {
		return ErrNotFound
	}
	key, e3 := uuid.Parse(in.RequestKey)
	if e3 != nil {
		return ErrValidation
	}
	return appError(a.Service.Publish(ctx, actor, gid, app.PublishInput{Version: in.Version, RequestKey: key}))
}
func (a AppAdapter) Close(ctx context.Context, u *User, id string) error {
	actor, e := a.actor(u)
	gid, e2 := uuid.Parse(id)
	if e != nil {
		return e
	}
	if e2 != nil {
		return ErrNotFound
	}
	return appError(a.Service.Close(ctx, actor, gid))
}
func (a AppAdapter) Assignment(ctx context.Context, u *User, id string) (Recipient, error) {
	actor, e := a.actor(u)
	gid, e2 := uuid.Parse(id)
	if e != nil {
		return Recipient{}, e
	}
	if e2 != nil {
		return Recipient{}, ErrNotFound
	}
	v, e := a.Service.MyAssignment(ctx, actor, gid)
	if e != nil {
		return Recipient{}, appError(e)
	}
	return Recipient{Recipient: Participant{Email: v.Email, DisplayName: v.Name}}, nil
}

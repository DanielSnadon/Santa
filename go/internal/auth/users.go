package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type UserStore struct{ DB *sql.DB }

func (s UserStore) Create(ctx context.Context, email, name, password string) (User, error) {
	if len(password) < 12 || len(password) > 256 {
		return User{}, errors.New("invalid password")
	}
	if len(name) == 0 || len(name) > 100 {
		return User{}, errors.New("invalid display name")
	}
	n := NormalizeEmail(email)
	if n == "" || len(n) > 256 {
		return User{}, errors.New("invalid email")
	}
	hash, e := HashIdentityV3(password)
	if e != nil {
		return User{}, e
	}
	u := User{ID: uuid.New(), Email: email, NormalizedEmail: n, UserName: email, NormalizedUserName: n, PasswordHash: hash, DisplayName: name, LockoutEnabled: true}
	_, e = s.DB.ExecContext(ctx, `INSERT INTO "AspNetUsers" ("Id","UserName","NormalizedUserName","Email","NormalizedEmail","EmailConfirmed","PasswordHash","SecurityStamp","ConcurrencyStamp","PhoneNumberConfirmed","TwoFactorEnabled","LockoutEnabled","AccessFailedCount","DisplayName") VALUES ($1,$2,$3,$4,$5,false,$6,$7,$8,false,false,true,0,$9)`, u.ID, u.UserName, u.NormalizedUserName, u.Email, u.NormalizedEmail, u.PasswordHash, uuid.NewString(), uuid.NewString(), u.DisplayName)
	if e != nil {
		return User{}, e
	}
	return u, nil
}
func (s UserStore) ByEmail(ctx context.Context, email string) (User, error) {
	var u User
	e := s.DB.QueryRowContext(ctx, `SELECT "Id","Email","NormalizedEmail","UserName","NormalizedUserName","PasswordHash","DisplayName","AccessFailedCount","LockoutEnd","LockoutEnabled" FROM "AspNetUsers" WHERE "NormalizedEmail"=$1`, NormalizeEmail(email)).Scan(&u.ID, &u.Email, &u.NormalizedEmail, &u.UserName, &u.NormalizedUserName, &u.PasswordHash, &u.DisplayName, &u.AccessFailedCount, &u.LockoutEnd, &u.LockoutEnabled)
	return u, e
}
func (s UserStore) CheckPassword(ctx context.Context, u User, password string) (bool, error) {
	if u.LockoutEnabled && u.LockoutEnd != nil && u.LockoutEnd.After(time.Now()) {
		return false, fmt.Errorf("locked out")
	}
	if VerifyIdentityV3(u.PasswordHash, password) {
		_, e := s.DB.ExecContext(ctx, `UPDATE "AspNetUsers" SET "AccessFailedCount"=0,"LockoutEnd"=NULL WHERE "Id"=$1`, u.ID)
		return e == nil, e
	}
	next := u.AccessFailedCount + 1
	if next >= 5 {
		t := time.Now().UTC().Add(15 * time.Minute)
		_, e := s.DB.ExecContext(ctx, `UPDATE "AspNetUsers" SET "AccessFailedCount"=$2,"LockoutEnd"=$3 WHERE "Id"=$1`, u.ID, next, t)
		return false, e
	}
	_, e := s.DB.ExecContext(ctx, `UPDATE "AspNetUsers" SET "AccessFailedCount"=$2 WHERE "Id"=$1`, u.ID, next)
	return false, e
}

func (s UserStore) ByID(ctx context.Context, id uuid.UUID) (User, error) {
	var u User
	e := s.DB.QueryRowContext(ctx, `SELECT "Id","Email","NormalizedEmail","UserName","NormalizedUserName","PasswordHash","DisplayName","AccessFailedCount","LockoutEnd","LockoutEnabled" FROM "AspNetUsers" WHERE "Id"=$1`, id).Scan(&u.ID, &u.Email, &u.NormalizedEmail, &u.UserName, &u.NormalizedUserName, &u.PasswordHash, &u.DisplayName, &u.AccessFailedCount, &u.LockoutEnd, &u.LockoutEnabled)
	return u, e
}

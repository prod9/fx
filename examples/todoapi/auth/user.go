package auth

import (
	"context"
	"time"

	"fx.prodigy9.co/data"
	. "fx.prodigy9.co/examples/todoapi/gen/todoapi/public/table"
	"fx.prodigy9.co/passwords"
	"github.com/go-jet/jet/v2/postgres"
)

type User struct {
	ID           int64     `json:"id" db:"id,users.id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

func GetUserByID(ctx context.Context, id int64) (*User, error) {
	user := &User{}
	if err := data.GetSQL(ctx, user, Users.
		SELECT(Users.AllColumns).
		WHERE(Users.ID.EQ(postgres.Int64(id))).
		LIMIT(1),
	); err != nil {
		return nil, err
	} else {
		return user, nil
	}
}

func GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	if err := data.GetSQL(ctx, user, Users.
		SELECT(Users.AllColumns).
		WHERE(Users.Username.EQ(postgres.String(username))).
		LIMIT(1),
	); err != nil {
		return nil, err
	} else {
		return user, nil
	}
}

func (u *User) SetPasswordHash(password string) error {
	hash, err := passwords.Hash(password)
	if err != nil {
		return err
	} else {
		u.PasswordHash = hash
		return nil
	}
}

func (u *User) ValidatePassword(password string) (bool, error) {
	err := passwords.Compare(u.PasswordHash, password)
	if err != nil {
		if passwords.IsMismatch(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

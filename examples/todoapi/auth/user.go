package auth

import (
	"context"
	"time"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/passwords"
)

type User struct {
	ID           int64     `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

func GetUserByID(ctx context.Context, id int64) (*User, error) {
	user := &User{}
	if err := data.Get(ctx, user, `SELECT * FROM users WHERE id = $1 LIMIT 1`, id); err != nil {
		return nil, err
	} else {
		return user, nil
	}
}

func GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	if err := data.Get(ctx, user, `SELECT * FROM users WHERE username = $1 LIMIT 1`, username); err != nil {
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

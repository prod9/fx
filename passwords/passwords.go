package passwords

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var ErrPasswordMismatch = fmt.Errorf("passwords: %w", bcrypt.ErrMismatchedHashAndPassword)

func IsMismatch(err error) bool {
	return errors.Is(err, ErrPasswordMismatch)
}

func Hash(pwd string) (string, error) {
	if buf, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost); err != nil {
		return "", err
	} else {
		return string(buf), nil // buf is already b64-encoded
	}
}

func Compare(hash, pwd string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))
	if err == nil {
		return nil
	} else if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrPasswordMismatch
	} else {
		return fmt.Errorf("passwords: %w", err)
	}
}

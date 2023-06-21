package config

import "errors"

var ErrEmpty = errors.New("Marker signifying empty value")

// Checks is the error is ErrEmpty signifying that a configuration has an empty value.
func IsEmpty(err error) bool {
	return errors.Is(err, ErrEmpty)
}

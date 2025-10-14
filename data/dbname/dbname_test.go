package dbname

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrom(t *testing.T) {
	barebones := "postgres:///mydb"
	name, err := From(barebones)
	require.NoError(t, err)
	require.Equal(t, "mydb", name)

	everything := "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
	name, err = From(everything)
	require.NoError(t, err)
	require.Equal(t, "mydb", name)
}

func TestSanitize(t *testing.T) {
	vanilla := Sanitize("HelloWorld123")
	require.Equal(t, "helloworld123", vanilla)

	slashed := Sanitize("/HelloWorld123")
	require.Equal(t, "helloworld123", slashed)

	long := Sanitize(strings.Repeat("X", 70))
	require.Equal(t, 63, len(long))
	require.Equal(t, strings.Repeat("x", 63), long)

	specials := Sanitize("!@#$%^&*()[]{};:,.<>?/\\|`~wooop")
	require.Equal(t, "wooop", specials)
}

func TestSet(t *testing.T) {
	barebones := "postgres:///olddb"
	updated, err := Set(barebones, "newdb")
	require.NoError(t, err)
	require.Equal(t, "postgres:///newdb", updated)

	basic := "postgres://localhost/olddb?sslmode=disable"
	updated, err = Set(basic, "newdb")
	require.NoError(t, err)
	require.Equal(t, "postgres://localhost/newdb?sslmode=disable", updated)

	full := "postgres://user:pass@localhost:5432/olddb?sslmode=disable"
	updated, err = Set(full, "newdb")
	require.NoError(t, err)
	require.Equal(t, "postgres://user:pass@localhost:5432/newdb?sslmode=disable", updated)
}

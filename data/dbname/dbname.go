package dbname

import (
	"net/url"
	"strings"
	"unicode"
)

func From(dbURL string) (string, error) {
	if u, err := url.Parse(dbURL); err != nil {
		return "", err
	} else {
		return strings.TrimPrefix(u.Path, "/"), nil
	}
}

// Sanitize strips out special chars, lowercases, and trims at 63 to meet PostgreSQL
// database name requirements.
func Sanitize(dbName string) string {
	var sb strings.Builder
	for i, r := range dbName {
		if i == 63 {
			break
		}
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			continue
		}

		sb.WriteRune(unicode.ToLower(r))
	}
	return sb.String()
}

func Set(dbURL, dbName string) (string, error) {
	if u, err := url.Parse(dbURL); err != nil {
		return "", err
	} else {
		u.Path = "/" + dbName
		return u.String(), nil
	}
}

func SetDefaultDB(dbURL string) (string, error) {
	return Set(dbURL, "postgres")
}

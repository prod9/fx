package config

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Constraint for the generic T used in all the vars, in case we ever need to change it or
// add more constraints
type _TType interface {
	comparable
}

type _Var interface {
	Name() string
	Aliases() []string

	getAny(*Source) any
	_var() // opaque marker
}

// Var allows each part of the application (and fx itself) to declare their own
// configuration variables and use it with the standard *Source without requiring changes
// to the config package directly.
type Var[T _TType] struct {
	name    string
	aliases []string
	defVal  T
	parse   func(string) (T, error)
}

func NewVar[T _TType](name string, defVal T, parse func(string) (T, error)) *Var[T] {
	v := &Var[T]{name: name, defVal: defVal, parse: parse}
	defaultSource.vars = append(defaultSource.vars, v)
	return v
}

// WithAlias adds older names the variable is also read from, tried in order after the
// canonical name. Call it at declaration:
//
//	var SentryDSNConfig = config.Str("SENTRY_DSN").WithAlias("API_SENTRY_DSN")
func (v *Var[T]) WithAlias(names ...string) *Var[T] {
	v.aliases = append(v.aliases, names...)
	return v
}

func (v Var[T]) Name() string   { return v.name }
func (v Var[T]) String() string { return v.name }

// Aliases returns the older names the variable is also read from, in lookup order.
func (v Var[T]) Aliases() []string { return v.aliases }

func (v *Var[T]) getAny(src *Source) any { return Get(src, v) }
func (_ Var[T]) _var()                   { /* marker */ }

func Str(name string) *Var[string]             { return NewVar(name, "", parseStr) }
func Int(name string) *Var[int]                { return NewVar(name, 0, parseInt) }
func Int64(name string) *Var[int64]            { return NewVar(name, 0, parseInt64) }
func URL(name string) *Var[*url.URL]           { return NewVar(name, nil, url.Parse) }
func Bool(name string) *Var[bool]              { return NewVar(name, false, strconv.ParseBool) }
func Duration(name string) *Var[time.Duration] { return NewVar(name, 0, time.ParseDuration) }

func StrDef(name, def string) *Var[string]            { return NewVar(name, def, parseStr) }
func IntDef(name string, def int) *Var[int]           { return NewVar(name, def, parseInt) }
func Int64Def(name string, def int64) *Var[int64]     { return NewVar(name, def, parseInt64) }
func URLDef(name string, def *url.URL) *Var[*url.URL] { return NewVar(name, def, url.Parse) }
func BoolDef(name string, def bool) *Var[bool]        { return NewVar(name, def, strconv.ParseBool) }
func DurationDef(name string, def time.Duration) *Var[time.Duration] {
	return NewVar(name, def, time.ParseDuration)
}

func parseStr(s string) (string, error) {
	if _s := strings.TrimSpace(s); _s == "" {
		return "", ErrEmpty
	} else {
		return s, nil
	}
}
func parseInt(s string) (int, error) {
	if _s := strings.TrimSpace(s); _s == "" {
		return 0, ErrEmpty
	} else if n, err := strconv.ParseInt(_s, 10, 0); err != nil {
		return 0, err
	} else {
		// since bitSize == 0 (see ParseInt doc) convert should be lossless.
		return int(n), nil
	}
}
func parseInt64(s string) (int64, error) {
	if _s := strings.TrimSpace(s); _s == "" {
		return 0, ErrEmpty
	} else if n, err := strconv.ParseInt(_s, 10, 64); err != nil {
		return 0, err
	} else {
		return n, nil
	}
}

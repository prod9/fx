package config

import (
	"net/url"
	"strconv"
	"strings"
)

// Constraint for the generic T used in all the vars, in case we ever need to change it or
// add more constraints
type _TType interface {
	comparable
}

type _Var interface {
	Name() string

	defaultAny() any
	parseAny(string) (any, error)
	_var() // opaque marker
}

// Var allows each part of the application (and fx itself) to declare their own
// configuration variables and use it with the standard *Source without requiring changes
// to the config package directly.
type Var[T _TType] struct {
	name   string
	defVal T
	parse  func(string) (T, error)
}

func NewVar[T _TType](name string, defVal T, parse func(string) (T, error)) *Var[T] {
	v := &Var[T]{name, defVal, parse}
	defaultSource.vars = append(defaultSource.vars, v)
	return v
}

func (v Var[T]) Name() string   { return v.name }
func (v Var[T]) String() string { return v.name }

func (v Var[T]) defaultAny() any                  { return v.defVal }
func (v Var[T]) parseAny(str string) (any, error) { return v.parse(str) }
func (_ Var[T]) _var()                            { /* marker */ }

func Str(name string) *Var[string]                    { return NewVar(name, "", parseStr) }
func StrDef(name, def string) *Var[string]            { return NewVar(name, def, parseStr) }
func Int(name string) *Var[int]                       { return NewVar(name, 0, parseInt) }
func IntDef(name string, def int) *Var[int]           { return NewVar(name, def, parseInt) }
func Int64(name string) *Var[int64]                   { return NewVar(name, 0, parseInt64) }
func Int64Def(name string, def int64) *Var[int64]     { return NewVar(name, def, parseInt64) }
func URL(name string) *Var[*url.URL]                  { return NewVar(name, nil, url.Parse) }
func URLDef(name string, def *url.URL) *Var[*url.URL] { return NewVar(name, def, url.Parse) }
func Bool(name string) *Var[bool]                     { return NewVar(name, false, strconv.ParseBool) }
func BoolDef(name string, def bool) *Var[bool]        { return NewVar(name, def, strconv.ParseBool) }

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

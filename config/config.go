package config

import (
	"fmt"
	"strings"
	"sync"

	// NOTE: Sadly, we cannot use fxlog here because it'd cause import cycle.
	// and `config` is intended to be the first package to initializes anyway.
	//
	// Perhaps we should make fxlog not dependent on `config`.
	"log"
)

var loggedAliases sync.Map

func GetAny(src *Source, v _Var) any {
	return v.getAny(src)
}

func Get[T _TType](src *Source, v *Var[T]) T {
	val, _ := GetOK(src, v)
	return val
}

// GetOK returns the variable's value and whether one was set. A missing, blank, or
// unparsable value, or a failed provider read, yields the default: a configuration
// problem must not prevent app startup (docs/spec/configuration.md).
func GetOK[T _TType](src *Source, v *Var[T]) (T, bool) {
	raw, ok := lookup(src, v)
	if !ok {
		return v.defVal, false
	}

	val, err := v.parse(raw)
	if IsEmpty(err) {
		return v.defVal, false
	} else if err != nil {
		log.Println(v.name+":", err)
		return v.defVal, false
	}
	return val, true
}

// lookup returns the first non-blank value among the variable's canonical name, then its
// aliases in order.
func lookup(src *Source, v _Var) (string, bool) {
	if raw, ok := read(src, v.Name()); ok {
		return raw, true
	}

	for _, alias := range v.Aliases() {
		if raw, ok := read(src, alias); ok {
			logAliasUse(v.Name(), alias)
			return raw, true
		}
	}
	return "", false
}

func read(src *Source, name string) (string, bool) {
	raw, ok, err := src.provider.Get(name)
	if err != nil {
		log.Println("config:", err)
		return "", false
	}

	raw = strings.TrimSpace(raw)
	return raw, ok && raw != ""
}

func logAliasUse(name, alias string) {
	if _, logged := loggedAliases.LoadOrStore(alias, struct{}{}); !logged {
		log.Printf("config: %s is read from its old name %s; rename it to %s", name, alias, name)
	}
}

// Set writes the value under the variable's canonical name only.
func Set[T _TType](src *Source, v *Var[T], val T) {
	err := src.provider.Set(v.Name(), fmt.Sprint(val))
	if err != nil {
		log.Println("config:", err)
	}
}

// SetDefault sets a new default value **globally** for the entire program overiding any
// default values specified during the Var's initialization.
func SetDefault[T _TType](v *Var[T], defVal T) {
	v.defVal = defVal
}

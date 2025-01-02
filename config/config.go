package config

import (
	"fmt"
	"log"
	"strings"
)

func GetAny(src *Source, v _Var) any {
	raw, ok, err := src.provider.Get(v.Name())
	if err != nil {
		log.Println("config:", err)
		return nil
	}

	raw = strings.TrimSpace(raw)
	if !ok || raw == "" {
		return v.defaultAny()
	} else if val, err := v.parseAny(raw); IsEmpty(err) {
		return v.defaultAny()
	} else if err != nil {
		log.Println(v.Name()+":", err)
		return v.defaultAny()
	} else {
		return val
	}
}

func GetOK[T _TType](src *Source, v *Var[T]) (T, bool) {
	raw, ok, err := src.provider.Get(v.Name())
	if err != nil {
		log.Println("config:", err)
		return v.defVal, false
	}

	raw = strings.TrimSpace(raw)
	if !ok || raw == "" {
		return v.defVal, false
	} else if val, err := v.parse(raw); IsEmpty(err) {
		return v.defVal, false
	} else if err != nil {
		log.Println(v.name+":", err)
		return v.defVal, false
	} else {
		return val, true
	}
}

func Get[T _TType](src *Source, v *Var[T]) T {
	raw, _, err := src.provider.Get(v.Name())
	if err != nil {
		log.Println("config:", err)
		return v.defVal
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return v.defVal
	} else if val, err := v.parse(raw); IsEmpty(err) {
		return v.defVal
	} else if err != nil {
		log.Println(v.name+":", err)
		return v.defVal
	} else {
		return val
	}
}

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

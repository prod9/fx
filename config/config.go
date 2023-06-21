package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func GetAny(src *Source, v _Var) any {
	env := strings.TrimSpace(os.Getenv(v.Name()))
	if env == "" {
		return v.defaultAny()
	} else if val, err := v.parseAny(env); IsEmpty(err) {
		return v.defaultAny()
	} else if err != nil {
		log.Println(v.Name()+":", err)
		return v.defaultAny()
	} else {
		return val
	}
}

func Get[T _TType](src *Source, v *Var[T]) T {
	env := strings.TrimSpace(os.Getenv(v.name))
	if env == "" {
		return v.defVal
	} else if val, err := v.parse(env); IsEmpty(err) {
		return v.defVal
	} else if err != nil {
		log.Println(v.name+":", err)
		return v.defVal
	} else {
		return val
	}
}

func Set[T _TType](src *Source, v *Var[T], val T) {
	os.Setenv(v.Name(), fmt.Sprint(val))
}

// SetDefault sets a new default value **globally** for the entire program overiding any
// default values specified during the Var's initialization.
func SetDefault[T _TType](v *Var[T], defVal T) {
	v.defVal = defVal
}

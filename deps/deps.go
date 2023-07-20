// Package deps
// stores providers for dependencies that will be mocked for tests
package deps

import (
	"fmt"
)

type Store struct {
	deps map[interface{}]interface{}
}

var defaultStore = &Store{deps: make(map[interface{}]interface{})}

// Set sets a provider, wont set if exists
func Set(key interface{}, provider func(params ...interface{}) interface{}) {
	if _, exists := defaultStore.deps[key]; exists {
		return
	}
	defaultStore.deps[key] = provider
}

// SetDirect sets a provider
func SetDirect(key interface{}, provider func(params ...interface{}) interface{}) {
	defaultStore.deps[key] = provider
}

// GetProvider get existing provider
func GetProvider(key interface{}) func(params ...interface{}) interface{} {
	if provider, exists := defaultStore.deps[key]; exists {
		p, _ := provider.(func(params ...interface{}) interface{})
		return p
	}
	return nil
}

// Execute executes a provider and returns the typed value
func Execute[T comparable](key interface{}, providerParams ...interface{}) T {
	providerVal, exists := defaultStore.deps[key]
	provider, typed := providerVal.(func(params ...interface{}) interface{})
	if !exists || !typed {
		panic(fmt.Sprintf("missing dependency with key: %T", key))
	}
	resultVal := provider(providerParams...)
	result, typed := resultVal.(T)
	if !typed {
		panic(fmt.Sprintf("couldn't cast dependency to target type: %T", new(T)))
	}
	return result
}

// Provider wraps a provider to add a parameter with a type for it
func Provider[T comparable](provider func(p T) interface{}) func(params ...interface{}) interface{} {
	return func(params ...interface{}) interface{} {
		if len(params) != 1 {
			panic("missing parameter")
		}
		ctx, ok := params[0].(T)
		if !ok {
			panic("wrong type")
		}
		return provider(ctx)
	}
}

package errutil

import (
	"strconv"
	"strings"
	"sync"
)

type taggedError struct {
	tag string
	err error
}

// Aggregator collects tagged errors from a batch of operations and reports them as a
// single error. Safe for concurrent use; the zero value is ready to use. Its Error
// renders one "tag: message" line per collected error.
type Aggregator struct {
	mutex sync.RWMutex

	any  bool
	errs []taggedError
}

// Aggregate runs action on every element of slice concurrently and collects the errors,
// each tagged by its index. Returns nil if all succeed, otherwise an *Aggregator (as an
// error) holding every failure.
func Aggregate[T any](slice []T, action func(int, T) error) error {
	wg, agg := sync.WaitGroup{}, &Aggregator{}

	for idx, item := range slice {
		wg.Add(1)
		go func(idx int, item T) {
			defer wg.Done()
			agg.Add(strconv.Itoa(idx), action(idx, item))
		}(idx, item)
	}

	wg.Wait()
	return agg.Return()
}

// AggregateWithTags is Aggregate where action supplies its own tag per element, to label
// errors by something more meaningful than the index.
func AggregateWithTags[T any](slice []T, action func(int, T) (tag string, err error)) error {
	wg, agg := sync.WaitGroup{}, &Aggregator{}

	for idx, item := range slice {
		wg.Add(1)
		go func(idx int, item T) {
			defer wg.Done()
			agg.Add(action(idx, item))
		}(idx, item)
	}

	wg.Wait()
	return agg.Return()
}

// Any reports whether at least one error has been collected.
func (a *Aggregator) Any() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return len(a.errs) > 0
}

// Len is the number of errors collected.
func (a *Aggregator) Len() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return len(a.errs)
}

// Add records err under tag and returns the aggregator for chaining. A nil err is a no-op,
// so callers can Add unconditionally.
func (a *Aggregator) Add(tag string, err error) *Aggregator {
	if err == nil {
		return a
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.errs = append(a.errs, taggedError{tag, err})
	return a
}

// Return yields nil when no errors were collected, otherwise the aggregator itself as an
// error — the idiomatic tail of an aggregation.
func (a *Aggregator) Return() error {
	if a.Any() {
		return a
	} else {
		return nil
	}
}

// Error implements error: one "tag: message" line per collected error.
func (a *Aggregator) Error() string {
	if a.Len() == 0 {
		return ""
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	sb := strings.Builder{}
	for _, te := range a.errs {
		sb.WriteString(te.tag)
		sb.WriteRune(':')
		sb.WriteRune(' ')
		sb.WriteString(te.err.Error())
		sb.WriteRune('\n')
	}
	return sb.String()
}

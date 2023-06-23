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

type Aggregator struct {
	mutex sync.RWMutex

	any  bool
	errs []taggedError
}

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

func (a *Aggregator) Any() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return len(a.errs) > 0
}

func (a *Aggregator) Len() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return len(a.errs)
}

func (a *Aggregator) Add(tag string, err error) *Aggregator {
	if err == nil {
		return a
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.errs = append(a.errs, taggedError{tag, err})
	return a
}

func (a *Aggregator) Return() error {
	if a.Any() {
		return a
	} else {
		return nil
	}
}

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

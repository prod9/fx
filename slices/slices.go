package slices

import (
	"golang.org/x/exp/constraints"
	"sort"
)

func In[T comparable](arr []T, find T) bool {
	for _, item := range arr {
		if item == find {
			return true
		}
	}
	return false
}

func Filter[T any](arr []T, f func(T) bool) []T {
	var result []T
	for _, item := range arr {
		if f(item) {
			result = append(result, item)
		}
	}
	return result
}

func Uniq[T constraints.Ordered](arr []T) []T {
	var result []T
	uniq := map[T]struct{}{}
	for _, item := range arr {
		if _, exists := uniq[item]; exists {
			continue
		}

		uniq[item] = struct{}{}
		result = append(result, item)
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

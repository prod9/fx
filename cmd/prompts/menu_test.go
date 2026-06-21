package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWindow(t *testing.T) {
	// list shorter than the viewport: show everything
	start, end := window(0, 3, 10)
	assert.Equal(t, [2]int{0, 3}, [2]int{start, end})

	// long list, cursor near the top: window pinned to start
	start, end = window(0, 20, 10)
	assert.Equal(t, [2]int{0, 10}, [2]int{start, end})

	// cursor mid-list: roughly centered
	start, end = window(10, 20, 10)
	assert.Equal(t, [2]int{5, 15}, [2]int{start, end})

	// cursor at the end: window pinned to the bottom
	start, end = window(19, 20, 10)
	assert.Equal(t, [2]int{10, 20}, [2]int{start, end})
}

func TestWrap(t *testing.T) {
	assert.Equal(t, 2, wrap(-1, 3)) // up past the top wraps to the bottom
	assert.Equal(t, 0, wrap(3, 3))  // down past the bottom wraps to the top
	assert.Equal(t, 1, wrap(1, 3))
	assert.Equal(t, 0, wrap(0, 0)) // empty list never panics
}

func TestIndexOf(t *testing.T) {
	opts := []string{"a", "b", "c"}
	assert.Equal(t, 1, indexOf(opts, "b"))
	assert.Equal(t, -1, indexOf(opts, "z"))
}

func TestPicked(t *testing.T) {
	opts := []string{"a", "b", "c"}
	assert.Equal(t, []string{"a", "c"}, picked(opts, []bool{true, false, true}))
	assert.Nil(t, picked(opts, []bool{false, false, false}))
}

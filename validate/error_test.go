package validate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError_Add(t *testing.T) {
	dummy := NewFieldError("dummy", "dummy", "dummy").(*FieldError)
	dummy2 := NewFieldError("dummy2", "dummy2", "dummy2").(*FieldError)

	var e *Error
	e = e.Add(dummy)
	require.NotNil(t, e)
	require.Len(t, e.Fields, 1)
	require.Equal(t, e.Fields["dummy"][0].Error(), dummy.Error())

	e = e.Add(dummy2)
	require.NotNil(t, e)
	require.Len(t, e.Fields, 2)
	require.Equal(t, e.Fields["dummy2"][0].Error(), dummy2.Error())
}

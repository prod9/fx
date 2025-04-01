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

func TestMulti(t *testing.T) {
	dummy := Multi(
		NewFieldError("dummy", "dummy", "dummy"),
		NewFieldError("dummy2", "dummy2", "dummy2"),
	)

	require.NotNil(t, dummy)
	require.IsType(t, &Error{}, dummy)

	err := dummy.(*Error)
	require.Len(t, err.Fields, 2)
	require.Equal(t, err.Fields["dummy"][0].Error(), "dummy: dummy")
	require.Equal(t, err.Fields["dummy2"][0].Error(), "dummy2: dummy2")
}

func TestMulti_WithGroup(t *testing.T) {
	dummy := Multi(
		NewFieldError("outer1", "msg1", "value1"),
		NewFieldError("outer2", "msg2", "value2"),
		Group("group1",
			NewFieldError("inner1", "msg3", "value3"),
			NewFieldError("inner2", "msg4", "value4"),
		),
		NewFieldError("outer3", "msg5", "value5"),
	)

	require.NotNil(t, dummy)
	require.IsType(t, &Error{}, dummy)

	err := dummy.(*Error)
	require.Len(t, err.Fields, 5)
	require.Equal(t, err.Fields["outer1"][0].Error(), "outer1: msg1")
	require.Equal(t, err.Fields["outer2"][0].Error(), "outer2: msg2")
	require.Equal(t, err.Fields["group1.inner1"][0].Error(), "group1.inner1: msg3")
	require.Equal(t, err.Fields["group1.inner2"][0].Error(), "group1.inner2: msg4")
	require.Equal(t, err.Fields["outer3"][0].Error(), "outer3: msg5")
}

func TestGroup(t *testing.T) {
	dummy := Group("ppppp",
		NewFieldError("dummy", "dummy", "dummy"),
		NewFieldError("dummy2", "dummy2", "dummy2"),
	)

	require.NotNil(t, dummy)
	require.IsType(t, &Error{}, dummy)

	err := dummy.(*Error)
	require.Len(t, err.Fields, 2)
	require.Equal(t, err.Fields["ppppp.dummy"][0].Error(), "ppppp.dummy: dummy")
	require.Equal(t, err.Fields["ppppp.dummy2"][0].Error(), "ppppp.dummy2: dummy2")
}

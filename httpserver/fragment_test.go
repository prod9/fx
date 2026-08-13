package httpserver

import (
	"testing"

	"fx.prodigy9.co/httpserver/controllers"
	"github.com/stretchr/testify/require"
)

func TestFragment_IsEmpty_SeesChildren(t *testing.T) {
	var (
		empty     = NewFragment(nil, nil)
		withCtrs  = NewFragment(nil, []controllers.Interface{controllers.Home{}})
		withChild = NewFragment(nil, nil).AddChild(withCtrs)
	)

	require.True(t, empty.IsEmpty())
	require.False(t, withCtrs.IsEmpty())
	require.False(t, withChild.IsEmpty(),
		"a fragment whose only content is a child's controllers is not empty")
	require.True(t, NewFragment(nil, nil).AddChild(NewFragment(nil, nil)).IsEmpty(),
		"empty children keep the fragment empty")
}

package auth

import "context"

type (
	userContextKey    struct{}
	sessionContextKey struct{}
)

func NewContextWithSession(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, s)
}
func SessionFromContext(ctx context.Context) *Session {
	if s, ok := ctx.Value(sessionContextKey{}).(*Session); ok {
		return s
	} else {
		return nil
	}
}

func NewContextWithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userContextKey{}, u)
}
func UserFromContext(ctx context.Context) *User {
	if u, ok := ctx.Value(userContextKey{}).(*User); ok {
		return u
	} else {
		return nil
	}
}

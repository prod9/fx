package auth

import (
	"net/http"
	"strings"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/errutil"
	"fx.prodigy9.co/httpserver/httperrors"
	"fx.prodigy9.co/httpserver/middlewares"
	"fx.prodigy9.co/httpserver/render"
)

var (
	ErrSessionExpired = errutil.NewCoded("session_expired", "session has expired", nil)

	_ middlewares.Interface = LoadSession
	_ middlewares.Interface = RequireSession
)

func GetBearerTokenFromRequest(req *http.Request) string {
	auth := req.Header.Get("Authorization")
	if len(auth) < 7 || !strings.EqualFold(auth[:7], "Bearer ") {
		return ""
	}

	return auth[7:]
}

func LoadSession(cfg *config.Source) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {

			token := GetBearerTokenFromRequest(req)
			sess, err := GetSessionByToken(req.Context(), token)
			switch {
			case data.IsNoRows(err):
				render.Error(resp, req, 403, httperrors.ErrUnauthorized)
				return
			case err != nil:
				render.Error(resp, req, 500, err)
				return
			case sess == nil:
				render.Error(resp, req, 403, httperrors.ErrUnauthorized)
				return
			case sess.IsExpired():
				render.Error(resp, req, 403, ErrSessionExpired)
				return
			}

			user, err := GetUserByID(req.Context(), sess.UserID)
			if err != nil { // if user not found we have a constraint error somewhere
				render.Error(resp, req, 500, err)
				return
			}

			ctx := req.Context()
			ctx = NewContextWithSession(ctx, sess)
			ctx = NewContextWithUser(ctx, user)
			next.ServeHTTP(resp, req.WithContext(ctx))

		})
	}
}

// superset of LoadSession, don't need to add LoadSession if using RequireSession already
func RequireSession(cfg *config.Source) func(http.Handler) http.Handler {
	loadSession := LoadSession(cfg)

	return func(next http.Handler) http.Handler {
		checkSession := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			sess := SessionFromContext(req.Context())
			if sess == nil {
				render.Error(resp, req, 403, httperrors.ErrUnauthorized)
				return
			}

			next.ServeHTTP(resp, req)
		})

		return loadSession(checkSession)
	}
}

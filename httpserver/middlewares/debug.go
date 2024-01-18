package middlewares

import (
	"io"
	"net/http"
	"os"

	"fx.prodigy9.co/config"
	"github.com/felixge/httpsnoop"
)

var (
	DebugRequestConfig  = config.Bool("DEBUG_REQUEST")
	DebugResponseConfig = config.Bool("DEBUG_RESPONSE")
)

type closeSegue struct {
	reader io.Reader
	closer io.Closer
}

func (r *closeSegue) Read(p []byte) (n int, err error) { return r.reader.Read(p) }
func (r *closeSegue) Close() error                     { return r.closer.Close() }

func DebugRequest(cfg *config.Source) func(http.Handler) http.Handler {
	debugOut := os.Stdout

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			if config.Get(cfg, DebugRequestConfig) {
				tee := io.TeeReader(req.Body, debugOut)
				req.Body = &closeSegue{reader: tee, closer: req.Body}
			}

			if config.Get(cfg, DebugResponseConfig) {
				resp = httpsnoop.Wrap(resp, httpsnoop.Hooks{
					Write: func(write httpsnoop.WriteFunc) httpsnoop.WriteFunc {
						return func(b []byte) (int, error) {
							debugOut.Write(b)
							return write(b)
						}
					},
				})
			}

			next.ServeHTTP(resp, req)
		})
	}
}

package render

import (
	"encoding/json"
	"io"
	"net/http"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/errutil"
	"fx.prodigy9.co/fxlog"
	"fx.prodigy9.co/httpserver/httperrors"
)

func Text(resp http.ResponseWriter, r *http.Request, text string) {
	resp.Header().Set("Content-Type", "text/plain")
	resp.WriteHeader(200)
	if _, err := resp.Write([]byte(text)); err != nil {
		Error(resp, r, http.StatusInternalServerError, err)
	}
}

func JSON(resp http.ResponseWriter, r *http.Request, obj interface{}) {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(200)
	if err := json.NewEncoder(resp).Encode(obj); err != nil {
		Error(resp, r, http.StatusInternalServerError, err)
	}
}

func Redirect(resp http.ResponseWriter, r *http.Request, target string) {
	http.Redirect(resp, r, target, http.StatusTemporaryRedirect)
}

// TODO: status code should be specified by the originating error since otherwise we'll
// have to make that decision in the controllers which breaks SRP.
func Error(resp http.ResponseWriter, r *http.Request, status int, err error) {
	// check for common global error types
	if data.IsNoRows(err) {
		Error(resp, r, 404, httperrors.ErrNotFound)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(status)

	errObj := errutil.Decorate(err)
	if err_ := json.NewEncoder(resp).Encode(errObj); err_ != nil {
		fxlog.Errorf("render: %s %s: json encoding error: %w",
			r.Method, r.RequestURI, err_)
	}
}

func FileTransfer(resp http.ResponseWriter, r *http.Request, filename string, reader io.Reader) {
	resp.Header().Set("Content-Transfer-Encoding", "binary")
	resp.Header().Set("Content-Disposition", "attachment; filename="+filename)
	resp.Header().Set("Content-Type", "application/octet-stream")
	resp.WriteHeader(200)
	if _, err := io.Copy(resp, reader); err != nil {
		Error(resp, r, http.StatusInternalServerError, err)
	}
}

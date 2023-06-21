package render

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"fx.prodigy9.co/data"
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

	errObj := decorateError(err)
	if err_ := json.NewEncoder(resp).Encode(errObj); err_ != nil {
		log.Printf("%s %s %s - %s\n",
			r.RemoteAddr, r.Method, r.RequestURI, err_.Error())
	}
}

func FileTransfer(resp http.ResponseWriter, r *http.Request, filename string, reader io.Reader) {
	resp.Header().Set("Content-Description", "File Transfer")
	resp.Header().Set("Content-Transfer-Encoding", "binary")
	resp.Header().Set("Content-Disposition", "attachment; filename="+filename)
	resp.Header().Set("Content-Type", "application/octet-stream")
	resp.WriteHeader(200)
	if _, err := io.Copy(resp, reader); err != nil {
		Error(resp, r, http.StatusInternalServerError, err)
	}
}

func decorateError(err error) interface{} {
	errObj := &struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "unknown",
		Message: err.Error(),
	}

	if code, ok := err.(interface{ Code() string }); ok {
		errObj.Code = code.Code()
	}

	return errObj
}

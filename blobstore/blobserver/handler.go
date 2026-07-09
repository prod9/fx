package blobserver

import (
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"fx.prodigy9.co/fxlog"
)

// Handler is a minimal, path-style object store over a local directory. It serves the
// verbs blobstore emits against S3 — GET/PUT/DELETE/HEAD on /{bucket}/{key} — persisting
// objects as plain files under the root. No auth or ACLs.
type Handler struct {
	root string
}

func NewHandler(root string) *Handler {
	return &Handler{root: root}
}

func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	full, ok := h.resolve(req.URL.Path)
	if !ok {
		http.Error(resp, "blobserver: invalid key", http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodPut:
		h.put(resp, req, full)
	case http.MethodGet:
		h.get(resp, req, full)
	case http.MethodHead:
		h.head(resp, full)
	case http.MethodDelete:
		h.delete(resp, full)
	default:
		http.Error(resp, "blobserver: method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) put(resp http.ResponseWriter, req *http.Request, full string) {
	defer req.Body.Close()

	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		h.fail(resp, "put", err)
		return
	}

	file, err := os.Create(full)
	if err != nil {
		h.fail(resp, "put", err)
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, req.Body); err != nil {
		h.fail(resp, "put", err)
		return
	}
}

func (h *Handler) get(resp http.ResponseWriter, req *http.Request, full string) {
	file, err := os.Open(full)
	if os.IsNotExist(err) {
		http.Error(resp, "blobserver: not found", http.StatusNotFound)
		return
	} else if err != nil {
		h.fail(resp, "get", err)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		h.fail(resp, "get", err)
		return
	}

	http.ServeContent(resp, req, filepath.Base(full), info.ModTime(), file)
}

func (h *Handler) head(resp http.ResponseWriter, full string) {
	info, err := os.Stat(full)
	if os.IsNotExist(err) {
		http.Error(resp, "blobserver: not found", http.StatusNotFound)
		return
	} else if err != nil {
		h.fail(resp, "head", err)
		return
	}

	resp.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	resp.Header().Set("Last-Modified", info.ModTime().UTC().Format(http.TimeFormat))
	resp.WriteHeader(http.StatusOK)
}

func (h *Handler) delete(resp http.ResponseWriter, full string) {
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		h.fail(resp, "delete", err)
		return
	}
	resp.WriteHeader(http.StatusNoContent)
}

// resolve maps a request path to an absolute file path guaranteed to sit under the
// root. Cleaning an absolutized key neutralizes any "../" before the join, so a
// hostile key can only ever land inside the root.
func (h *Handler) resolve(urlPath string) (string, bool) {
	key := path.Clean("/" + strings.TrimPrefix(urlPath, "/"))
	full := filepath.Join(h.root, filepath.FromSlash(key))

	rel, err := filepath.Rel(h.root, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return full, true
}

func (h *Handler) fail(resp http.ResponseWriter, op string, err error) {
	fxlog.Log("blobserver error", fxlog.String("op", op), fxlog.String("error", err.Error()))
	http.Error(resp, "blobserver: "+op+" failed", http.StatusInternalServerError)
}

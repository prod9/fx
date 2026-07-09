package blobserver

import (
	"encoding/xml"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// placeholderETag stands in for a real content hash. The reconciliation sweep only
// reads object keys, and minio-go does not validate the ETag on a list response.
const placeholderETag = `"00000000000000000000000000000000"`

// listBucketResult is the S3 ListObjectsV2 (list-type=2) response shape minio-go's
// ListObjects call expects. Everything fits one page — a local dev store is small, so
// there is no continuation-token pagination.
type listBucketResult struct {
	XMLName     xml.Name     `xml:"ListBucketResult"`
	Xmlns       string       `xml:"xmlns,attr"`
	Name        string       `xml:"Name"`
	Prefix      string       `xml:"Prefix"`
	KeyCount    int          `xml:"KeyCount"`
	MaxKeys     int          `xml:"MaxKeys"`
	IsTruncated bool         `xml:"IsTruncated"`
	Contents    []listObject `xml:"Contents"`
}

type listObject struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

func (h *Handler) list(resp http.ResponseWriter, req *http.Request) {
	bucketDir, ok := h.resolve(req.URL.Path)
	if !ok {
		http.Error(resp, "blobserver: invalid bucket", http.StatusBadRequest)
		return
	}
	bucket := strings.Trim(req.URL.Path, "/")
	prefix := req.URL.Query().Get("prefix")

	contents := h.collect(bucketDir, prefix)
	result := listBucketResult{
		Xmlns:       "http://s3.amazonaws.com/doc/2006-03-01/",
		Name:        bucket,
		Prefix:      prefix,
		KeyCount:    len(contents),
		MaxKeys:     1000,
		IsTruncated: false,
		Contents:    contents,
	}

	resp.Header().Set("Content-Type", "application/xml")
	io.WriteString(resp, xml.Header)
	if err := xml.NewEncoder(resp).Encode(result); err != nil {
		h.fail(resp, "list", err)
	}
}

// collect walks the bucket directory and returns one entry per file whose key carries
// the prefix. A missing bucket directory yields an empty listing, not an error.
func (h *Handler) collect(bucketDir, prefix string) []listObject {
	var contents []listObject
	filepath.WalkDir(bucketDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(bucketDir, p)
		if err != nil {
			return nil
		}
		key := filepath.ToSlash(rel)
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}
		contents = append(contents, listObject{
			Key:          key,
			LastModified: info.ModTime().UTC().Format(time.RFC3339Nano),
			ETag:         placeholderETag,
			Size:         info.Size(),
			StorageClass: "STANDARD",
		})
		return nil
	})
	return contents
}

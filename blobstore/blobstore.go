package blobstore

import (
	"context"
	"net/http"

	"fx.prodigy9.co/config"
	"github.com/minio/minio-go/v7"
)

var (
	// ex: s3://key:secret@endpoint/bucket
	StorageURLConfig = config.Str("STORAGE_URL")
	DefaultClient    = NewClient(nil)
)

func PresignedGetURL(ctx context.Context, key string, options ...Option) (string, error) {
	return DefaultClient.PresignedGetURL(ctx, key, options...)
}
func PresignedPutURL(ctx context.Context, key string, options ...Option) (string, error) {
	return DefaultClient.PresignedPutURL(ctx, key, options...)
}
func DeleteObject(ctx context.Context, key string) error {
	return DefaultClient.DeleteObject(ctx, key)
}
func ObjectExists(ctx context.Context, key string) (bool, error) {
	return DefaultClient.ObjectExists(ctx, key)
}

func IsNotFound(err error) bool {
	return minio.ToErrorResponse(err).StatusCode == http.StatusNotFound
}

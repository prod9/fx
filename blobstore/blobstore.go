package blobstore

import (
	"context"
	"fx.prodigy9.co/config"
)

var (
	// ex: s3://key:secret@endpoint/bucket
	StorgeURLConfig = config.Str("STORAGE_URL")
	DefaultClient   = NewClient(nil)
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
func ForceDeleteObject(ctx context.Context, key string) error {
	return DefaultClient.ForceDeleteObject(ctx, key)
}

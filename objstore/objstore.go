package objstore

import (
	"context"
	"fmt"
	"log"
	"time"

	"fx.prodigy9.co/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageClient struct {
	accessKey string
	secretKey string
	url       string
	bucket    string
	client    *minio.Client
}

var (
	StorageAccessKeyConfig = config.Str("STORAGE_ACCESS_KEY")
	StorageSecretKeyConfig = config.Str("STORAGE_SECRET_KEY")
	StorageURLConfig       = config.Str("STORAGE_URL")
	StorageBucketConfig    = config.Str("STORAGE_BUCKET_NAME")

	DefaultClient *StorageClient
)

func NewClient(cfg *config.Source) *StorageClient {
	return newS3Client(cfg)
}

func getDefaultClient() {
	if DefaultClient == nil {
		DefaultClient = newS3Client(config.Configure())
	}
}

func newS3Client(cfg *config.Source) *StorageClient {
	if cfg == nil {
		cfg = config.Configure()
	}

	accessKey := config.Get(cfg, StorageAccessKeyConfig)
	secretKey := config.Get(cfg, StorageSecretKeyConfig)
	url := config.Get(cfg, StorageURLConfig)
	bucket := config.Get(cfg, StorageBucketConfig)

	client, err := minio.New(url, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Println("objstore: ", err)
	}

	return &StorageClient{
		accessKey: accessKey,
		secretKey: secretKey,
		url:       url,
		bucket:    bucket,
		client:    client,
	}
}

func (s *StorageClient) WithBucket(bucket string) *StorageClient {
	clone := *s
	clone.bucket = bucket
	return &clone
}

func (s *StorageClient) tryGetClient() *StorageClient {
	if s.client == nil {
		getDefaultClient()
		return DefaultClient
	}
	return s
}

func (s *StorageClient) PresignedGetURL(objectKey string, age time.Duration) (string, error) {
	s = s.tryGetClient()

	presignedURL, err := s.client.PresignedGetObject(context.Background(), s.bucket, objectKey, age, nil)
	if err != nil {
		return "", fmt.Errorf("objstore: %w", err)
	}

	return presignedURL.String(), nil
}

func (s *StorageClient) PresignedPutURL(objectKey string, age time.Duration) (string, error) {
	s = s.tryGetClient()

	presignedURL, err := s.client.PresignedPutObject(context.Background(), s.bucket, objectKey, age)
	if err != nil {
		return "", fmt.Errorf("objstore: %w", err)
	}

	return presignedURL.String(), nil
}

func (s *StorageClient) DeleteObject(objectKey string) error {
	s = s.tryGetClient()

	if err := s.client.RemoveObject(context.Background(), s.bucket, objectKey, minio.RemoveObjectOptions{ForceDelete: true}); err != nil {
		return fmt.Errorf("objstore: %w", err)
	}

	return nil
}

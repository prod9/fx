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

type Client struct {
	bucket string
	cfg    *config.Source
	client *minio.Client
}

var DefaultClient = &Client{}

var (
	StorageAccessKeyConfig = config.Str("STORAGE_ACCESS_KEY")
	StorageSecretKeyConfig = config.Str("STORAGE_SECRET_KEY")
	StorageEndpointConfig  = config.Str("STORAGE_ENDPOINT")
	StorageBucketConfig    = config.Str("STORAGE_BUCKET")
)

func NewClient(cfg *config.Source) *Client {
	return newS3Client(cfg)
}

func newS3Client(cfg *config.Source) *Client {
	if cfg == nil {
		cfg = config.Configure()
	}

	accessKey := config.Get(cfg, StorageAccessKeyConfig)
	secretKey := config.Get(cfg, StorageSecretKeyConfig)
	url := config.Get(cfg, StorageEndpointConfig)
	bucket := config.Get(cfg, StorageBucketConfig)

	client, err := minio.New(url, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Println("objstore: ", err)
	}

	return &Client{
		bucket: bucket,
		client: client,
	}
}

func (s *Client) tryGetClient() *Client {
	if s.client == nil {
		var err error
		if DefaultClient.client, err = DefaultClient.getMinio(); err != nil {
			log.Println("objstore: ", err)
		}
		s = DefaultClient
	}
	return s
}

func (c *Client) getMinio() (*minio.Client, error) {
	if c.cfg == nil {
		c.cfg = config.Configure()
	}

	if c.client == nil {
		c = newS3Client(c.cfg)
		DefaultClient.bucket = c.bucket
	}

	return c.client, nil
}

func (s *Client) WithBucket(bucket string) *Client {
	clone := *s
	clone.bucket = bucket
	return &clone
}

func (s *Client) PresignedGetURL(objectKey string, age time.Duration) (string, error) {
	s = s.tryGetClient()
	presignedURL, err := s.client.PresignedGetObject(context.Background(), s.bucket, objectKey, age, nil)
	if err != nil {
		return "", fmt.Errorf("objstore: %w", err)
	}
	return presignedURL.String(), nil
}

func (s *Client) PresignedPutURL(objectKey string, age time.Duration) (string, error) {
	s = s.tryGetClient()
	presignedURL, err := s.client.PresignedPutObject(context.Background(), s.bucket, objectKey, age)
	if err != nil {
		return "", fmt.Errorf("objstore: %w", err)
	}
	return presignedURL.String(), nil
}

func (s *Client) DeleteObject(objectKey string) error {
	s = s.tryGetClient()
	if err := s.client.RemoveObject(context.Background(), s.bucket, objectKey, minio.RemoveObjectOptions{ForceDelete: true}); err != nil {
		return fmt.Errorf("objstore: %w", err)
	}
	return nil
}

func PresignedGetURL(objectKey string, age time.Duration) (string, error) {
	defaultClient := DefaultClient.tryGetClient()
	presignedURL, err := defaultClient.client.PresignedGetObject(context.Background(), defaultClient.bucket, objectKey, age, nil)
	if err != nil {
		return "", fmt.Errorf("objstore: %w", err)
	}
	return presignedURL.String(), nil
}

func PresignedPutURL(objectKey string, age time.Duration) (string, error) {
	defaultClient := DefaultClient.tryGetClient()
	presignedURL, err := defaultClient.client.PresignedPutObject(context.Background(), defaultClient.bucket, objectKey, age)
	if err != nil {
		return "", fmt.Errorf("objstore: %w", err)
	}
	return presignedURL.String(), nil
}

func DeleteObject(objectKey string, age time.Duration) error {
	defaultClient := DefaultClient.tryGetClient()
	if err := defaultClient.client.RemoveObject(context.Background(), defaultClient.bucket, objectKey, minio.RemoveObjectOptions{ForceDelete: true}); err != nil {
		return fmt.Errorf("objstore: %w", err)
	}
	return nil
}

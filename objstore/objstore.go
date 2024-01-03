package objstore

import (
	"context"
	"fmt"
	"time"

	"fx.prodigy9.co/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	MinioAccessKeyConfig = config.Str("MINIO_ACCESS_KEY")
	MinioSecretKeyConfig = config.Str("MINIO_SECRET_KEY")
	MinioEndpointConfig  = config.Str("MINIO_ENDPOINT")

	client *minio.Client
)

func connect(cfg *config.Source) (*minio.Client, error) {
	if client != nil {
		return client, nil
	}

	minioAccessKey := config.Get(cfg, MinioAccessKeyConfig)
	minioSecretKey := config.Get(cfg, MinioSecretKeyConfig)
	minioEndpoint := config.Get(cfg, MinioEndpointConfig)

	var err error
	client, err = minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: true,
	})
	if err != nil {
		client = nil
		return nil, err
	}

	return client, nil
}

func GeneratePresignedGetURL(bucket, objectKey string, expiry time.Duration) (string, error) {
	client, err := connect(config.Configure())
	if err != nil {
		return "", fmt.Errorf("failed to initialize cloud client: %v", err)
	}

	presignedURL, err := client.PresignedGetObject(context.Background(), bucket, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return presignedURL.String(), nil
}

func GeneratePresignedPutURL(bucket, objectKey string, expiry time.Duration) (string, error) {
	client, err := connect(config.Configure())
	if err != nil {
		return "", fmt.Errorf("failed to initialize cloud client: %v", err)
	}

	presignedURL, err := client.PresignedPutObject(context.Background(), bucket, objectKey, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return presignedURL.String(), nil
}

func DeleteObject(bucket, objectKey string) error {
	client, err := connect(config.Configure())
	if err != nil {
		return fmt.Errorf("failed to initialize cloud client: %v", err)
	}

	if err := client.RemoveObject(context.Background(), bucket, objectKey, minio.RemoveObjectOptions{ForceDelete: true}); err != nil {
		return fmt.Errorf("failed to delete object: %v", err)
	}

	return nil
}

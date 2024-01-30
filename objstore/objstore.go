package objstore

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"fx.prodigy9.co/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	// ex: s3://key:secret@endpoint/bucket
	StorgeURLConfig = config.Str("STORAGE_URL")
	DefaultClient   = NewClient(nil)
)

func PresignedGetURL(ctx context.Context, key string, age time.Duration) (string, error) {
	return DefaultClient.PresignedGetURL(ctx, key, age)
}
func PresignedPutURL(ctx context.Context, key string, age time.Duration) (string, error) {
	return DefaultClient.PresignedPutURL(ctx, key, age)
}
func DeleteObject(ctx context.Context, key string) error {
	return DefaultClient.DeleteObject(ctx, key)
}
func ForceDeleteObject(ctx context.Context, key string) error {
	return DefaultClient.ForceDeleteObject(ctx, key)
}

type Client struct {
	mutex sync.RWMutex

	cfg    *config.Source
	client *minio.Client
	bucket string
}

func NewClient(cfg *config.Source) *Client {
	if cfg == nil {
		cfg = config.Configure()
	}

	return &Client{
		cfg:    cfg,
		client: nil,
	}
}

func (s *Client) PresignedGetURL(ctx context.Context, key string, age time.Duration) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if client, err := s.tryGetMinio(); err != nil {
		return "", err
	} else if u, err := client.PresignedGetObject(ctx, s.bucket, key, age, nil); err != nil {
		return "", fmt.Errorf("objstore: %w", err)
	} else {
		return u.String(), nil
	}
}

func (s *Client) PresignedPutURL(ctx context.Context, key string, age time.Duration) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if client, err := s.tryGetMinio(); err != nil {
		return "", err
	} else if u, err := client.PresignedPutObject(ctx, s.bucket, key, age); err != nil {
		return "", fmt.Errorf("objstore: %w", err)
	} else {
		return u.String(), nil
	}
}

func (s *Client) DeleteObject(ctx context.Context, key string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	opts := minio.RemoveObjectOptions{}
	if client, err := s.tryGetMinio(); err != nil {
		return err
	} else if err := client.RemoveObject(ctx, s.bucket, key, opts); err != nil {
		return fmt.Errorf("objstore: %w", err)
	} else {
		return nil
	}
}

func (s *Client) ForceDeleteObject(ctx context.Context, key string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	opts := minio.RemoveObjectOptions{ForceDelete: true}
	if client, err := s.tryGetMinio(); err != nil {
		return err
	} else if err := client.RemoveObject(ctx, s.bucket, key, opts); err != nil {
		return fmt.Errorf("objstore: %w", err)
	} else {
		return nil
	}
}

func (s *Client) tryGetMinio() (*minio.Client, error) {
	if client := s.getMinio(); client != nil {
		return client, nil
	} else if err := s.initMinio(); err != nil {
		return nil, err
	} else {
		return s.tryGetMinio() // should be initialized by this point
	}
}

func (s *Client) getMinio() *minio.Client {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.client
}

func (s *Client) initMinio() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s3url, err := url.Parse(config.Get(s.cfg, StorgeURLConfig))
	if err != nil {
		return fmt.Errorf("objstore: %w", err)
	}

	accessKey := s3url.User.Username()
	secretKey, hasSecretKey := s3url.User.Password()
	endpoint := s3url.Host
	bucket := strings.TrimPrefix(s3url.Path, "/")
	if accessKey == "" && !hasSecretKey {
		return errors.New("objstore: access key or secret key not configured")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
	if err != nil {
		return fmt.Errorf("objstore: %w", err)
	}

	s.client = client
	s.bucket = bucket
	return nil
}

package blobstore

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"fx.prodigy9.co/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

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

func (s *Client) PresignedGetURL(ctx context.Context, key string, options ...Option) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	opts := *defaults
	opts.apply(options...)

	if client, err := s.tryGetMinio(); err != nil {
		return "", err
	} else if u, err := client.PresignedGetObject(ctx, s.bucket, key, opts.age, nil); err != nil {
		return "", fmt.Errorf("blobstore: %w", err)
	} else {
		return u.String(), nil
	}
}

func (s *Client) PresignedPutURL(ctx context.Context, key string, options ...Option) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	opts := *defaults
	opts.apply(options...)

	headers := http.Header{}
	if opts.contentType != "" {
		headers.Set("Content-Type", opts.contentType)
	}
	if opts.contentLength > 0 {
		headers.Set("Content-Length", strconv.FormatInt(opts.contentLength, 10))
	}

	if client, err := s.tryGetMinio(); err != nil {
		return "", err
	} else if u, err := client.PresignHeader(ctx,
		"PUT", s.bucket, key,
		opts.age, nil, headers,
	); err != nil {
		return "", fmt.Errorf("blobstore: %w", err)
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
		return fmt.Errorf("blobstore: %w", err)
	} else {
		return nil
	}
}

// ObjectExists reports whether an object exists at key. A missing object is (false, nil);
// only a transport or access failure yields a non-nil error, so a nil error is an
// authoritative present/absent answer the reconciliation sweep can act on.
func (s *Client) ObjectExists(ctx context.Context, key string) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	client, err := s.tryGetMinio()
	if err != nil {
		return false, err
	}

	if _, err := client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{}); err != nil {
		if minio.ToErrorResponse(err).StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, fmt.Errorf("blobstore: %w", err)
	}
	return true, nil
}

// secureFromScheme maps a STORAGE_URL scheme to minio's TLS toggle. Whitelisted:
// s3 and https talk TLS (live endpoints); http targets a plaintext local blobserver.
func secureFromScheme(scheme string) (bool, error) {
	switch scheme {
	case "s3", "https":
		return true, nil
	case "http":
		return false, nil
	default:
		return false, fmt.Errorf("unsupported STORAGE_URL scheme %q", scheme)
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

	s3url, err := url.Parse(config.Get(s.cfg, StorageURLConfig))
	if err != nil {
		return fmt.Errorf("blobstore: %w", err)
	}

	accessKey := s3url.User.Username()
	secretKey, hasSecretKey := s3url.User.Password()
	endpoint := s3url.Host
	bucket := strings.TrimPrefix(s3url.Path, "/")
	if accessKey == "" && !hasSecretKey {
		return errors.New("blobstore: access key or secret key not configured")
	}

	secure, err := secureFromScheme(s3url.Scheme)
	if err != nil {
		return fmt.Errorf("blobstore: %w", err)
	}

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	}
	if !secure {
		// A plaintext endpoint is the local blobserver, which ignores auth and
		// region. Pin a region so minio-go skips the GetBucketLocation lookup the
		// server does not implement.
		opts.Region = "us-east-1"
	}

	client, err := minio.New(endpoint, opts)
	if err != nil {
		return fmt.Errorf("blobstore: %w", err)
	}

	s.client = client
	s.bucket = bucket
	return nil
}

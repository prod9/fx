package files

import (
	"context"
	"strconv"
	"time"

	"fx.prodigy9.co/blobstore"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/page"
)

type (
	FileKey struct {
		Kind      string `json:"kind" db:"kind"`
		OwnerType string `json:"owner_type" db:"owner_type"`
		OwnerID   int64  `json:"owner_id" db:"owner_id"`
		ID        int64  `json:"id" db:"id"`
	}

	File struct {
		ID        int64  `json:"id" db:"id"`
		Kind      string `json:"kind" db:"kind"`
		OwnerType string `json:"owner_type" db:"owner_type"`
		OwnerID   int64  `json:"owner_id" db:"owner_id"`

		OriginalName  string    `json:"original_name" db:"original_name"`
		ContentType   string    `json:"content_type" db:"content_type"`
		ContentLength int64     `json:"content_length" db:"content_length"`
		CreatedAt     time.Time `json:"created_at" db:"created_at"`
	}
)

func (f *File) PresignedGetURL(ctx context.Context) (string, error) {
	return blobstore.PresignedGetURL(ctx, f.RemotePath(), blobstore.WithAge(linkAge))
}

func (f *File) PresignedPutURL(ctx context.Context) (string, error) {
	return blobstore.PresignedPutURL(ctx, f.RemotePath(),
		blobstore.WithAge(linkAge),
		blobstore.WithContentType(f.ContentType),
		blobstore.WithContentLength(f.ContentLength),
	)
}

func (f *File) RemotePath() string {
	return f.OwnerType + "/" +
		f.Kind + "/" +
		strconv.FormatInt(f.OwnerID, 10) + "/" +
		strconv.FormatInt(f.ID, 10)
}

func GetUniqueFile(ctx context.Context, key FileKey) (*File, error) {
	sql := `
	SELECT * FROM files
	WHERE kind = $1 AND owner_type = $2 AND owner_id = $3
	ORDER BY created_at DESC
	LIMIT 1`

	file := &File{}
	if err := data.Get(ctx, file, sql, key.Kind, key.OwnerType, key.OwnerID); err != nil {
		return nil, err
	} else {
		return file, nil
	}
}

func GetFileByID(ctx context.Context, key FileKey) (*File, error) {
	sql := `
	SELECT * FROM files
	WHERE kind = $1 AND owner_type = $2 AND owner_id = $3 AND id = $4
	LIMIT 1`

	file := &File{}
	if err := data.Get(
		ctx, file, sql,
		key.Kind, key.OwnerType, key.OwnerID, key.ID,
	); err != nil {
		return nil, err
	} else {
		return file, nil
	}
}

func ListFiles(ctx context.Context, key FileKey, pm page.Meta) (*page.Page[*File], error) {
	sql := `
	SELECT * FROM files
	WHERE kind = $1 AND owner_type = $2 AND owner_id = $3
	ORDER BY created_at ASC`

	files := &page.Page[*File]{}
	if err := page.Select(
		ctx, files, pm, sql,
		key.Kind, key.OwnerType, key.OwnerID,
	); err != nil {
		return nil, err
	} else {
		return files, nil
	}
}

func DestroyFile(ctx context.Context, key FileKey) (*File, error) {
	sql := `
	DELETE FROM files
	WHERE kind = $1 AND owner_type = $2 AND owner_id = $3 AND id = $4
	RETURNING *`

	file := &File{}
	if err := data.Get(
		ctx, file, sql,
		key.Kind, key.OwnerType,
		key.OwnerID, key.ID,
	); err != nil {
		return nil, err
	}

	if err := blobstore.DeleteObject(ctx, file.RemotePath()); err != nil {
		return file, err
	}
	return file, nil
}

func DestroyUniqueFile(ctx context.Context, key FileKey) (*File, error) {
	sql := `
	DELETE FROM files
	WHERE kind = $1 AND owner_type = $2 AND owner_id = $3
	RETURNING *`

	file := &File{}
	if err := data.Get(
		ctx, file, sql,
		key.Kind, key.OwnerType,
		key.OwnerID,
	); err != nil {
		return nil, err
	}

	if err := blobstore.DeleteObject(ctx, file.RemotePath()); err != nil {
		return file, err
	}
	return file, nil
}

package files

import (
	"context"
	"time"

	"fx.prodigy9.co/blobstore"
)

type UploadInfo struct {
	FileInfo  *File  `json:"file_info"`
	UploadURL string `json:"upload_url"`
}

func UploadInfoFromFile(ctx context.Context, client *blobstore.Client, file *File, age time.Duration) (UploadInfo, error) {
	url, err := file.PresignedPutURL(ctx, client, age)
	if err != nil {
		return UploadInfo{}, err
	}

	return UploadInfo{
		FileInfo:  file,
		UploadURL: url,
	}, nil
}

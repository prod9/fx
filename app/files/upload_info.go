package files

import "context"

type UploadInfo struct {
	FileInfo  *File  `json:"file_info"`
	UploadURL string `json:"upload_url"`
}

func UploadInfoFromFile(ctx context.Context, file *File) (UploadInfo, error) {
	url, err := file.PresignedPutURL(ctx)
	if err != nil {
		return UploadInfo{}, err
	}

	return UploadInfo{
		FileInfo:  file,
		UploadURL: url,
	}, nil
}

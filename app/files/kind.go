package files

import "slices"

// defaultMaxSize caps an upload when a Kind leaves MaxSize zero. 256 MiB comfortably
// covers a full-resolution photo (incl. 48MP ProRAW) from a current phone.
const defaultMaxSize = 256 << 20

type Kind struct {
	Name         string
	Multiple     bool
	OwnerType    string
	ContentTypes []string

	// MaxSize caps the declared upload size in bytes. Zero uses defaultMaxSize; a
	// negative value means unlimited; a positive value is an explicit cap.
	MaxSize int64
}

func (k Kind) isValidContentType(contentType string) bool {
	return slices.Contains(k.ContentTypes, contentType)
}

func (k Kind) allowsSize(size int64) bool {
	max := k.MaxSize
	if max == 0 {
		max = defaultMaxSize
	}
	if max < 0 {
		return true
	}
	return size <= max
}

func (k Kind) key(ownerID, id int64) FileKey {
	return FileKey{
		Kind:      k.Name,
		OwnerType: k.OwnerType,
		OwnerID:   ownerID,
		ID:        id,
	}
}

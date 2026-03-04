package files

import "slices"

type Kind struct {
	Name         string
	Multiple     bool
	OwnerType    string
	ContentTypes []string
}

func (k Kind) isValidContentType(contentType string) bool {
	return slices.Contains(k.ContentTypes, contentType)
}

func (k Kind) key(ownerID, id int64) FileKey {
	return FileKey{
		Kind:      k.Name,
		OwnerType: k.OwnerType,
		OwnerID:   ownerID,
		ID:        id,
	}
}

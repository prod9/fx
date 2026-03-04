package files

import (
	"context"
	"mime"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/validate"
)

type CreateFile struct {
	Kind    Kind  `json:"-"` // specified through controller option
	OwnerID int64 `json:"-"`

	OriginalName  string `json:"original_name"`
	ContentType   string `json:"content_type"`
	ContentLength int64  `json:"content_length"`
}

func (c *CreateFile) Validate() error {
	return validate.Multi(
		c.validateKind(),
		c.validateContentType(),
		validate.Positive("owner_id", c.OwnerID),
		validate.Required("original_name", c.OriginalName),
		validate.Positive("content_length", c.ContentLength),
	)
}
func (c *CreateFile) validateKind() error {
	return validate.Group("kind",
		validate.Required("name", c.Kind.Name),
		validate.Required("owner_type", c.Kind.OwnerType),
	)
}
func (c *CreateFile) validateContentType() error {
	if c.ContentType == "" {
		return validate.Required("content_type", "")
	}
	if mediaType, _, err := mime.ParseMediaType(c.ContentType); err != nil {
		return validate.NewFieldError("content_type", "invalid", c.ContentType)
	} else if !c.Kind.isValidContentType(mediaType) {
		return validate.NewFieldError("content_type", "unsupported", c.ContentType)
	}
	return nil
}

func (c *CreateFile) Execute(ctx context.Context, out any) error {
	sql := `
	INSERT INTO files (
		kind, owner_type, owner_id,
		original_name, content_type, content_length
	)
	VALUES (
		$1, $2, $3,
		$4, $5, $6
	)
	RETURNING *`

	return data.Get(ctx, out, sql,
		c.Kind.Name, c.Kind.OwnerType, c.OwnerID,
		c.OriginalName, c.ContentType, c.ContentLength,
	)
}

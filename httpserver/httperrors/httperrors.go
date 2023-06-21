package httperrors

var (
	ErrNotFound     = &errImpl{"not_found", "not found"}
	ErrUnauthorized = &errImpl{"unauthorized", "unauthorized"}
	ErrInternal     = &errImpl{"internal", "internal server error"}
	ErrBadRequest   = &errImpl{"bad_request", "bad request"}
)

type errImpl struct {
	code    string
	message string
}

func (i *errImpl) Code() string  { return i.code }
func (i *errImpl) Error() string { return i.message }

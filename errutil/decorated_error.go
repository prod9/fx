package errutil

type decoratedErr struct {
	inner error

	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *decoratedErr) Error() string { return e.Message }
func (e *decoratedErr) Unwrap() error { return e.inner }

func NewCoded(code, msg string, data any) error {
	return &decoratedErr{inner: nil, Code: code, Message: msg, Data: data}
}

func Decorate(err error) error {
	if err == nil {
		return nil
	}

	outerr := &decoratedErr{inner: err, Message: err.Error()}
	if code, ok := err.(interface{ Code() string }); ok {
		outerr.Code = code.Code()
	}
	if data, ok := err.(interface{ ErrorData() any }); ok {
		outerr.Data = data.ErrorData()
	}
	return outerr
}

func WithCode(err error, code string) error {
	if err == nil {
		return err
	}

	outerr := &decoratedErr{inner: err, Message: err.Error(), Code: code}
	if data, ok := err.(interface{ ErrorData() any }); ok {
		outerr.Data = data.ErrorData()
	}
	return outerr
}

func WithData(err error, data any) error {
	if err == nil {
		return err
	}

	outerr := &decoratedErr{inner: err, Message: err.Error(), Data: data}
	if code, ok := err.(interface{ Code() string }); ok {
		outerr.Code = code.Code()
	}
	return outerr
}

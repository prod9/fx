package blobstore

import "time"

type options struct {
	age           time.Duration
	contentType   string
	contentLength int64
}

var defaults = &options{
	age: 5 * time.Minute,

	contentType:   "",
	contentLength: 0,
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func (o *options) setDefaults() {
	if o.age == 0 {
		o.age = 5 * time.Minute
	}
}

type Option func(o *options)

func WithAge(age time.Duration) Option {
	return func(o *options) { o.age = age }
}
func WithContentType(contentType string) Option {
	return func(o *options) { o.contentType = contentType }
}
func WithContentLength(contentLength int64) Option {
	return func(o *options) { o.contentLength = contentLength }
}

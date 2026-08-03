package pubsub

// Channel is an immutable descriptor — a name and a phantom T — holding no live state;
// every operation routes through the package (Publish/Subscribe), mirroring config.Get.
// Declare one at package scope next to its consumer, named after the var that holds it:
//
//	var OrdersChanged = pubsub.NewChannel[OrderEvent]("orders_changed")
//
// Signal-only channels, where the fact of the event is the whole message, use struct{}:
//
//	var CacheFlushed = pubsub.NewChannel[struct{}]("cache_flushed")
type Channel[T any] struct {
	name string
}

// NewChannel returns a typed channel descriptor. It panics on an invalid name — a name
// collision or malformed identifier is a program-construction error that must surface at
// startup, not a runtime condition to handle.
func NewChannel[T any](name string) Channel[T] {
	if err := validateName(name); err != nil {
		panic(err.Error())
	}
	return Channel[T]{name: name}
}

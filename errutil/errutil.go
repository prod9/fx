package errutil

import (
	"fmt"
)

func Wrap(name string, errptr *error) {
	if *errptr != nil {
		*errptr = fmt.Errorf(name+": %w", *errptr)
	}
}

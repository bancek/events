package events

import (
	"errors"
	"reflect"
)

func UnwrapAll(err error) error {
	if err == nil {
		return nil
	}
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

// GetCause unwraps the error and returns the original error if it is not the
// same as the passed error.
func GetCause(err error) (error, bool) {
	if err == nil {
		return nil, false
	}
	cause := UnwrapAll(err)
	if !reflect.TypeOf(err).Comparable() || !reflect.TypeOf(cause).Comparable() {
		if reflect.DeepEqual(err, cause) {
			return nil, false
		}
		return cause, true
	}
	if cause == err {
		return nil, false
	}
	return cause, true
}

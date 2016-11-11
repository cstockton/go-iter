package iter

import (
	"fmt"
	"reflect"
)

// Functions that are not worth an import dependency are in here, they come from
// another package of mine, go-refutil and are well tested.

// indirect will perform recursive indirection on the given value. It should
// never panic and will return a value unless indirection is impossible due to
// infinite recursion in cases like `type Element *Element`.
func indirect(value interface{}) interface{} {

	// Just to be safe, recursion should not be possible but I may be
	// missing an edge case.
	for {

		val := reflect.ValueOf(value)
		if !val.IsValid() || val.Kind() != reflect.Ptr {
			// Value is not a pointer.
			return value
		}

		res := reflect.Indirect(val)
		if !res.IsValid() || !res.CanInterface() {
			// Invalid value or can't be returned as interface{}.
			return value
		}

		// Test for a circular type.
		if res.Kind() == reflect.Ptr && val.Pointer() == res.Pointer() {
			return value
		}

		// Next round.
		value = res.Interface()
	}
}

// recoverFn will attempt to execute f, if f return a non-nil error it will be
// returned. If f panics this function will attempt to recover() and return a
// error instead.
func recoverFn(f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch T := r.(type) {
			case error:
				err = T
			default:
				err = fmt.Errorf("panic: %v", r)
			}
		}
	}()
	err = f()
	return
}

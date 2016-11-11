// Package iter provides primitives for walking arbitrary data structures.
package iter

import (
	"fmt"
	"reflect"
)

var (
	defaultIter   = Iter{}
	defaultWalker = dfsWalker{Iterator: &defaultIter}
)

// Iterator is a basic interface for iterating elements of a structured type. It
// serves as backing for other traversal methods. Iterators are safe for use by
// multiple Go routines, though the underlying values received in the iteration
// functions may not be.
type Iterator interface {
	IterChan(val reflect.Value, f func(seq int, ch reflect.Value) error) error
	IterMap(val reflect.Value, f func(key, val reflect.Value) error) error
	IterSlice(val reflect.Value, f func(idx int, val reflect.Value) error) error
	IterStruct(val reflect.Value,
		f func(field reflect.StructField, val reflect.Value) error) error
}

// NewIter returns a new Iter.
func NewIter() Iterator {
	return &Iter{}
}

// NewRecoverIter returns the given iterator wrapped so that it will not panic
// under any circumstance, instead returning the panic as an error.
func NewRecoverIter(it Iterator) Iterator {
	return &recoverIter{it}
}

// Iter it a basic implementation of Iterator. It is possible for these methods
// to panic under certain circumstances. If you want to disable panics write it
// in a iter.NewrecoverFnIter(Iter{}). Performance cost for Visiting slices is
// negligible relative to iteration using range. About 3x slower for slices with
// less than 1k elements, for slices  with more than 2k elements it will be
// around 2x as slow. Iterating maps is only 2-3 times slower for small maps of
// 100-1k elements. When you start to go above that its deficiencies will start
// to take a linear tax on runtime proportianite to the number of elements. It's
// roughly 100x slower at 32k elements. This is because it loads all map keys
// into memory (implementation of reflect.MapKeys).
type Iter struct {
	ChanRecv          bool
	ChanBlock         bool
	ExcludeAnonymous  bool
	ExcludeUnexported bool
}

// IterMap will visit each key and value of a map.
func (it Iter) IterMap(val reflect.Value, f func(
	key, val reflect.Value) error) error {
	kind := val.Kind()
	if reflect.Map != kind {
		return fmt.Errorf("expected map kind, not %s", kind)
	}
	for _, key := range val.MapKeys() {
		element := val.MapIndex(key)
		if err := f(key, element); err != nil {
			return err
		}
	}
	return nil
}

// IterSlice will visit each element of an array or slice. Extending the length
// of an Array or Slice during iteration may panic.
func (it Iter) IterSlice(val reflect.Value, f func(
	idx int, val reflect.Value) error) error {
	kind := val.Kind()
	if reflect.Slice != kind && kind != reflect.Array {
		return fmt.Errorf("expected array or slice kind, not %s", kind)
	}
	l := val.Len()
	for i := 0; i < l; i++ {
		element := val.Index(i)
		if err := f(i, element); err != nil {
			return err
		}
	}
	return nil
}

// IterStruct will visit each field and value in a struct.
func (it Iter) IterStruct(val reflect.Value, f func(
	field reflect.StructField, val reflect.Value) error) error {
	kind := val.Kind()
	if reflect.Struct != kind {
		return fmt.Errorf("expected struct kind, not %s", kind)
	}
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.Anonymous && it.ExcludeAnonymous {
			continue
		}
		if len(field.PkgPath) > 0 && it.ExcludeUnexported {
			continue
		}
		element := val.Field(i)
		if err := f(field, element); err != nil {
			return err
		}
	}
	return nil
}

// IterChan will try to receive values from the given channel only if ChanRecv
// is set to true. If ChanBlock is true IterChan will walk values until the
// channel has been clocked, otherwise it will use the behavior described in
// the reflect packages Value.TryRecv. This means when setting ChanBlock it is
// up to the caller to close the channel to prevent a dead lock. A sequential
// counter for this iterations receives is returned for parity with structured
// types.
func (it Iter) IterChan(val reflect.Value, f func(
	seq int, recv reflect.Value) error) error {
	if !it.ChanRecv {
		return nil
	}
	kind := val.Kind()
	if reflect.Chan != kind {
		return fmt.Errorf("expected chan kind, not %s", kind)
	}
	var (
		recv reflect.Value
		ok   bool
	)
	i := -1
	for {
		if it.ChanBlock {
			recv, ok = val.Recv()
		} else {
			recv, ok = val.TryRecv()
		}
		if !ok {
			return nil
		}
		i++
		if err := f(i, recv); err != nil {
			return err
		}
	}
}

type recoverIter struct {
	Iterator
}

func (it recoverIter) IterMap(val reflect.Value, f func(
	key, val reflect.Value) error) (err error) {
	return recoverFn(func() error {
		return it.Iterator.IterMap(val, f)
	})
}

func (it recoverIter) IterSlice(val reflect.Value, f func(
	idx int, val reflect.Value) error) (err error) {
	return recoverFn(func() error {
		return it.Iterator.IterSlice(val, f)
	})
}

func (it recoverIter) IterStruct(val reflect.Value, f func(
	field reflect.StructField, val reflect.Value) error) (err error) {
	return recoverFn(func() error {
		return it.Iterator.IterStruct(val, f)
	})
}

func (it recoverIter) IterChan(val reflect.Value, f func(
	seq int, recv reflect.Value) error) (err error) {
	return recoverFn(func() error {
		return it.Iterator.IterChan(val, f)
	})
}

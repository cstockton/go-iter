package iter

import "reflect"

// Walk will recursively walk the given interface value as long as an error does
// not occur. The pair func will be given a interface value for each value
// visited during walking and is expected to return an error if it thinks the
// traversal should end. A nil value and error is given to the walk func if an
// inaccessible value (can't reflect.Interface()) is found.
//
// Walk is called on each element of maps, slices and arrays. If the underlying
// iterator is configured for channels it receives until one fails. Channels
// should probably be avoided as ranging over them is more concise.
func Walk(value interface{}, f func(el Pair) error) error {
	return defaultWalker.Walk(value, f)
}

// A Walker is used to perform a full traversal of each child value of any Go
// type. Each implementation may use their own algorithm for traversal, giving
// no guarantee for the order each element is visited in.
type Walker interface {
	Walk(value interface{}, f func(el Pair) error) error
}

// NewWalker returns a new Walker backed by the given Iterator. It will use a
// basic dfs traversal and will not visit items that can not be converted to an
// interface.
func NewWalker(iterator Iterator) Walker {
	return &dfsWalker{Iterator: iterator}
}

type dfsWalker struct {
	Iterator
}

func (w dfsWalker) Walk(value interface{}, f func(el Pair) error) error {
	root := &pair{
		key: nil,
		val: value,
		pnt: nil,
		err: nil,
	}
	return w.walk(root, f)
}

func (w dfsWalker) walk(el Pair, f func(Pair) error) error {
	in := reflect.ValueOf(indirect(el.Val()))

	switch in.Kind() {
	case reflect.Slice, reflect.Array:
		return w.IterSlice(in, w.seqVisitFunc(el, f))
	case reflect.Struct:
		return w.IterStruct(in, w.structVisitFunc(el, f))
	case reflect.Chan:
		return w.IterChan(in, w.seqVisitFunc(el, f))
	case reflect.Map:
		return w.IterMap(in, w.mapVisitFunc(el, f))
	default:
		return f(el)
	}
}

type structVisitFn func(field reflect.StructField, value reflect.Value) error

func (w dfsWalker) structVisitFunc(el Pair, f func(Pair) error) structVisitFn {
	return func(s reflect.StructField, v reflect.Value) error {
		if !v.IsValid() || !v.CanInterface() {
			return nil
		}
		return w.walk(&pair{s, v.Interface(), el, nil}, f)
	}
}

type seqVisitFunc func(idx int, value reflect.Value) error

func (w dfsWalker) seqVisitFunc(el Pair, f func(Pair) error) seqVisitFunc {
	return func(idx int, v reflect.Value) error {
		if !v.IsValid() || !v.CanInterface() {
			return nil
		}
		return w.walk(&pair{idx, v.Interface(), el, nil}, f)
	}
}

type mapVisitFunc func(key, value reflect.Value) error

func (w dfsWalker) mapVisitFunc(el Pair, f func(Pair) error) mapVisitFunc {
	return func(k, v reflect.Value) error {
		if !k.IsValid() || !v.IsValid() ||
			!k.CanInterface() || !v.CanInterface() {
			return nil
		}
		return w.walk(&pair{k.Interface(), v.Interface(), el, nil}, f)
	}
}

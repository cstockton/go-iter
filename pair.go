package iter

import (
	"fmt"
	"reflect"
)

// Pair represents a dyadic pair of values visited by a Walker. The root Pair
// will have no Parent() and it is permissible for a Pair to contain only a Key
// or Val. This means any of this interfaces types with nil zero values could be
// nil regardless if Err() is non-nil. The Err() field may be populated by the
// Walker or propagated from a user returned error.
//
// Caveats:
//
//   Since channels do not have a sequential number to represent a location
//   within a finite space a simple numeric counter relative to the first
//   receive operation within the current block instead. This means that future
//   calls to the same channel could return a identical sequence number.
//
// Example of elements:
//
//   []int{1, 2} -> []Pair{ {0, 1}, {1, 2} }
//   map[str]int{"a": 1, "b": 2} -> []Pair{ {"a", 1}, {"b", 2} }
//
// Structs:
//
//   Key() will contain a reflect.StructFieldvalue, which may be anonymous or
//   unexported.
//   Val() will contain the associated fields current value.
//
// Slices, Arrays and Channels:
//
//   Key() will contain a int type representing the elements location within the
//   sequence.
//   Val() will contain the element value located at Key().
//
// Maps:
//
//   Key() will be a interface value of the maps key used to access the Val().
//   Val() will be a interface value of the value located at Key().
//
type Pair interface {

	// Err will return any error associated with the retrieval of this Pair.
	Err() error

	// Depth returns how many structured elements this Pair is nested within.
	Depth() int

	// Parent will return the parent Pair this Pair is associated to. For
	// example if this Pair belongs to a map type, the parent would contain a
	// Pair with Kind() reflect.Map and a reflect.Value of the map Value. If
	// this is the top most element then it will have no Parent.
	Parent() Pair

	// Key will return the key associated with this value.
	Key() interface{}

	// Val will return the value associated with this key.
	Val() interface{}

	// Pair returns the key and value for this Pair.
	Pair() (interface{}, interface{})
}

type pair struct {
	key interface{}
	val interface{}
	pnt Pair
	err error
}

// NewPair returns a new key-value pair to be used by Walkers.
func NewPair(parent Pair, key, value interface{}, err error) Pair {
	return &pair{
		key: key,
		val: value,
		pnt: parent,
		err: err,
	}
}

func (pr *pair) Err() error {
	return pr.err
}

func (pr *pair) Depth() int {
	if pr.pnt == nil {
		return 0
	}
	return 1 + pr.pnt.Depth()
}

func (pr *pair) Parent() Pair {
	return pr.pnt
}

func (pr *pair) Key() interface{} {
	return pr.key
}

func (pr *pair) Val() interface{} {
	return pr.val
}

func (pr *pair) Pair() (key, val interface{}) {
	return pr.key, pr.val
}

func (pr *pair) String() string {
	var k string
	if v, ok := pr.key.(reflect.StructField); ok {
		k = v.Name
	} else {
		k = fmt.Sprintf("%v", pr.key)
	}
	return fmt.Sprintf("Pair{(%T) %.12s => %.12s (%T)}",
		pr.key, k, fmt.Sprintf("%v", pr.val), pr.val)
}

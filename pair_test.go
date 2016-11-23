package iter

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestPairInterface(t *testing.T) {
	var _ Pair = (*pair)(nil)
}

func TestPair(t *testing.T) {
	tok := func(t *testing.T, pr Pair) {
		if err := pr.Err(); err != nil {
			t.Fatalf("pair returned error: %v", err)
		}
	}
	rk := "rkey"
	rv := "rval"
	root := NewPair(nil, rk, rv, nil)
	tok(t, root)

	t.Run("String", func(t *testing.T) {
		exp := `Pair{(string) rkey => rval (string)}`
		if got := fmt.Sprintf("%v", root); got != exp {
			t.Errorf("String() failed:\n  exp: %#v\n  got: %#v", exp, got)
		}
	})

	t.Run("String", func(t *testing.T) {
		type tstruct struct {
			Str string
		}
		typ := reflect.TypeOf(tstruct{"foo"})
		pr := NewPair(nil, typ.Field(0), "foo", nil)
		exp := `Pair{(reflect.StructField) Str => foo (string)}`
		if got := fmt.Sprintf("%v", pr); got != exp {
			t.Errorf("String() failed:\n  exp: %#v\n  got: %#v", exp, got)
		}
	})
	t.Run("Pair", func(t *testing.T) {
		k, v := root.Pair()
		if !reflect.DeepEqual(rk, k) {
			t.Errorf("DeepEqual failed Pair() key:\n  exp: %#v\n  got: %#v", rk, k)
		}
		if !reflect.DeepEqual(rv, v) {
			t.Errorf("DeepEqual failed Pair() value:\n  exp: %#v\n  got: %#v", rv, v)
		}
	})

	t.Run("Err", func(t *testing.T) {
		err := errors.New("exp err")
		pr := &pair{err: err}
		pr.err = err
		if got := pr.Err(); !reflect.DeepEqual(err, got) {
			t.Errorf("DeepEqual failed pr.Err():\n  exp: %#v\n  got: %#v", err, got)
		}
	})

	t.Run("Parent", func(t *testing.T) {
		k := "key"
		pr := NewPair(root, k, nil, nil)
		if got := pr.Parent(); !reflect.DeepEqual(root, got) {
			t.Errorf("DeepEqual failed pr.Parent():\n  exp: %#v\n  got: %#v", k, got)
		}
	})

	t.Run("Depth", func(t *testing.T) {
		exp := 100
		var root Pair
		for i := 0; i <= exp; i++ {
			root = NewPair(root, i, fmt.Sprintf("Val%d", i), nil)
			tok(t, root)
			if got := root.Depth(); got != i {
				t.Errorf("failed pr.Depth():\n  exp: %#v\n  got: %#v", i, got)
			}
		}
		if got := root.Depth(); got != exp {
			t.Errorf("failed pr.Depth():\n  exp: %#v\n  got: %#v", exp, got)
		}
	})

	t.Run("Key", func(t *testing.T) {
		k := "key"
		pr := NewPair(nil, k, nil, nil)
		if got := pr.Key(); !reflect.DeepEqual(k, got) {
			t.Errorf("DeepEqual failed pr.Key():\n  exp: %#v\n  got: %#v", k, got)
		}
	})

	t.Run("Value", func(t *testing.T) {
		v := "val"
		pr := NewPair(nil, v, nil, nil)
		if got := pr.Key(); !reflect.DeepEqual(v, got) {
			t.Errorf("DeepEqual failed pr.Val():\n  exp: %#v\n  got: %#v", v, got)
		}
	})
}

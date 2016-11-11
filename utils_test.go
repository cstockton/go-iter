package iter

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

var (
	nilValues = []interface{}{
		(*interface{})(nil), (**interface{})(nil), (***interface{})(nil),
		(func())(nil), (*func())(nil), (**func())(nil), (***func())(nil),
		(chan int)(nil), (*chan int)(nil), (**chan int)(nil), (***chan int)(nil),
		([]int)(nil), (*[]int)(nil), (**[]int)(nil), (***[]int)(nil),
		(map[int]int)(nil), (*map[int]int)(nil), (**map[int]int)(nil),
		(***map[int]int)(nil),
	}
)

func TestIndirect(t *testing.T) {
	type testindirectCircular *testindirectCircular
	teq := func(t testing.TB, exp, got interface{}) {
		if !reflect.DeepEqual(exp, got) {
			t.Errorf("DeepEqual failed:\n  exp: %#v\n  got: %#v", exp, got)
		}
	}

	t.Run("Basic", func(t *testing.T) {
		int64v := int64(123)
		int64vp := &int64v
		int64vpp := &int64vp
		int64vppp := &int64vpp
		int64vpppp := &int64vppp
		teq(t, indirect(int64v), int64v)
		teq(t, indirect(int64vp), int64v)
		teq(t, indirect(int64vpp), int64v)
		teq(t, indirect(int64vppp), int64v)
		teq(t, indirect(int64vpppp), int64v)
	})
	t.Run("Nils", func(t *testing.T) {
		for _, n := range nilValues {
			indirect(n)
		}
	})
	t.Run("Circular", func(t *testing.T) {
		var circular testindirectCircular
		circular = &circular
		teq(t, indirect(circular), circular)
	})
}

func TestRecoverFn(t *testing.T) {
	t.Run("CallsFunc", func(t *testing.T) {
		var called bool

		err := recoverFn(func() error {
			called = true
			return nil
		})
		if err != nil {
			t.Error("expected no error in recoverFn()")
		}
		if !called {
			t.Error("Expected recoverFn() to call func")
		}
	})
	t.Run("PropagatesError", func(t *testing.T) {
		err := fmt.Errorf("expect this error")
		rerr := recoverFn(func() error {
			return err
		})
		if err != rerr {
			t.Error("expected recoverFn() to propagate")
		}
	})
	t.Run("PropagatesPanicError", func(t *testing.T) {
		err := fmt.Errorf("expect this error")
		rerr := recoverFn(func() error {
			panic(err)
		})
		if err != rerr {
			t.Error("Expected recoverFn() to propagate")
		}
	})
	t.Run("PropagatesRuntimeError", func(t *testing.T) {
		err := recoverFn(func() error {
			sl := []int{}
			_ = sl[0]
			return nil
		})
		if err == nil {
			t.Error("expected runtime error to propagate")
		}
		if _, ok := err.(runtime.Error); !ok {
			t.Error("expected runtime error to retain type type")
		}
	})
	t.Run("PropagatesString", func(t *testing.T) {
		exp := "panic: string type panic"
		rerr := recoverFn(func() error {
			panic("string type panic")
		})
		if exp != rerr.Error() {
			t.Errorf("expected recoverFn() to return %v, got: %v", exp, rerr)
		}
	})
}

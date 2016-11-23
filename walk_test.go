package iter

import (
	"bytes"
	"container/ring"
	"fmt"
	"strings"
	"testing"
)

type TestTree struct {
	At       int
	Children []*TestTree
}

func (t *TestTree) Walk(parent Pair, f func(el Pair) error) error {
	return f(NewPair(parent, "foo", "bar", nil))
}

func newTestTree() *TestTree {
	return new(TestTree).fill(0, 4, 4)
}

func (t *TestTree) Dump() string {
	var buf bytes.Buffer
	space := strings.Repeat("  ", t.At)
	buf.WriteString(fmt.Sprintf("%sTree(%d)\n", space, t.At))
	for _, c := range t.Children {
		buf.WriteString(fmt.Sprintf("%s%v", space, c.Dump()))
	}
	return buf.String()
}

func (t *TestTree) String() string {
	return fmt.Sprintf("Tree(%d)\n", t.At)
}

func (t *TestTree) fill(d, c, max int) *TestTree {
	t.At = d
	d++
	if d >= max {
		return t
	}
	t.Children = make([]*TestTree, c)
	for i := 0; i < c; i++ {
		t.Children[i] = new(TestTree).fill(d, c, max)
	}
	return t
}

func TestDfsWalker(t *testing.T) {
	type testDfsWalker struct {
		Head       string
		Child      *testDfsWalker
		MapEnter   string
		Map        map[string]int
		SliceEnter string
		Slice      []string
		ChanEnter  string
		Chan       chan string
		Tail       string
	}
	tid := func(ident string, depth int) string {
		return fmt.Sprintf("%s|%d", ident, depth)
	}
	trnew := func(pnt *testDfsWalker, i int) *testDfsWalker {
		tr := &testDfsWalker{
			Head:     tid("Head", i),
			Child:    pnt,
			MapEnter: tid("MapEnter|1", i),
			Map: map[string]int{
				tid("Map", i):   i * 2,
				tid("Map", i+1): i * 2,
				tid("Map", i+2): i * 2},
			SliceEnter: tid("SliceEnter", i),
			Slice:      []string{tid("Slice", i*10), tid("Slice", i*20), tid("Slice", i*30)},
			ChanEnter:  tid("ChanEnter", i),
			Chan:       make(chan string, 3),
			Tail:       tid("Tail", i),
		}
		tr.Chan <- tid("Chan", i*10)
		tr.Chan <- tid("Chan", i*20)
		tr.Chan <- tid("Chan", i*30)
		return tr
	}
	tr := trnew(nil, 4)
	for i := 3; i > 0; i-- {
		tr = trnew(tr, i)
	}

	w := NewWalker(&Iter{ChanRecv: true})
	t.Run("Walk", func(t *testing.T) {
		fns := []func(value interface{}, f func(el Pair) error) error{
			Walk, w.Walk,
		}
		for _, fn := range fns {
			err := fn(tr, func(el Pair) error {
				return nil
			})
			if err != nil {
				t.Fatalf("expected nil err, got: %v", err)
			}
		}
	})

	t.Run("NoCircular", func(t *testing.T) {
		fns := []func(value interface{}, f func(el Pair) error) error{
			Walk, w.Walk,
		}
		for _, fn := range fns {
			i := 0
			err := fn(ring.New(10), func(el Pair) error {
				i++
				return nil
			})
			if err != nil {
				t.Fatalf("expected nil err, got: %v", err)
			}
			if i != 1 {
				t.Fatalf("expected exactly 1 visit, got: %v", i)
			}
		}
	})

	t.Run("NilChan", func(t *testing.T) {
		type privateString string
		w := NewWalker(invalidIter{})
		i := 0
		ch := make(chan string, 1)
		ch <- "foo"

		err := w.Walk(ch, func(el Pair) error {
			i++
			return nil
		})
		if err != nil {
			t.Fatalf("expected nil err, got: %v", err)
		}
		err = w.Walk(map[int]string{1: "123"}, func(el Pair) error {
			i++
			return nil
		})
		if err != nil {
			t.Fatalf("expected nil err, got: %v", err)
		}
		err = w.Walk([]string{"123"}, func(el Pair) error {
			i++
			return nil
		})
		if err != nil {
			t.Fatalf("expected nil err, got: %v", err)
		}
		err = w.Walk(struct{ name string }{"123"}, func(el Pair) error {
			i++
			return nil
		})
		if err != nil {
			t.Fatalf("expected nil err, got: %v", err)
		}
		if i != 0 {
			t.Fatalf("expected exactly 0 visits, got: %v", i)
		}
	})
}

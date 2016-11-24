package iter_test

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	iter "github.com/cstockton/go-iter"
)

func Example() {

	// Iter will walk any structured type, chans, slices, maps, structs.
	type Package struct {
		Name     string
		Synopsis string

		// Iter ascends into child structured types as well.
		Parent *Package
	}

	// Iter will visit each key -> value pair. For structs the key will be the
	// reflect.StructField, for maps the element key and the sequence integer for
	// arrays and, slices.
	pkgs := []Package{
		{"io", "Package io provides basic interfaces to I/O primitives.", &Package{
			"ioutil", "Package ioutil implements some I/O utility functions.", nil}},
		{"hash", "Package hash provides interfaces for hash functions.", nil},
		{"flag", "Package flag implements command-line flag parsing.", nil}}

	fmt.Println("Packages:")
	iter.Walk(pkgs, func(el iter.Pair) error {
		fmt.Printf("  %v\n", el)
		return nil
	})

	// Output:
	// Packages:
	//   Pair{(reflect.StructField) Name => io (string)}
	//   Pair{(reflect.StructField) Synopsis => Package io p (string)}
	//   Pair{(reflect.StructField) Name => ioutil (string)}
	//   Pair{(reflect.StructField) Synopsis => Package iout (string)}
	//   Pair{(reflect.StructField) Parent => <nil> (*iter_test.Package)}
	//   Pair{(reflect.StructField) Name => hash (string)}
	//   Pair{(reflect.StructField) Synopsis => Package hash (string)}
	//   Pair{(reflect.StructField) Parent => <nil> (*iter_test.Package)}
	//   Pair{(reflect.StructField) Name => flag (string)}
	//   Pair{(reflect.StructField) Synopsis => Package flag (string)}
	//   Pair{(reflect.StructField) Parent => <nil> (*iter_test.Package)}
}

func ExampleWalk() {

	var res []string
	m := map[int]string{1: "a", 2: "b", 3: "c"}

	err := iter.Walk(m, func(el iter.Pair) error {
		res = append(res, fmt.Sprintf("%v", el))
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	sort.Strings(res) // for test determinism
	for _, v := range res {
		fmt.Println(v)
	}

	// Output:
	// Pair{(int) 1 => a (string)}
	// Pair{(int) 2 => b (string)}
	// Pair{(int) 3 => c (string)}
}

func ExampleWalk_errors() {

	v := []interface{}{"a", "b", []string{"c", "d"}}
	err := iter.Walk(v, func(el iter.Pair) error {
		// check for errors
		if err := el.Err(); err != nil {
			return err
		}

		// Halt iteration by returning an error.
		if el.Depth() > 1 {
			return errors.New("Stopping this walk.")
		}

		fmt.Println(el)
		return nil
	})
	if err == nil {
		fmt.Println("expected an error")
	}

	// Output:
	// Pair{(int) 0 => a (string)}
	// Pair{(int) 1 => b (string)}
}

func Example_recursion() {

	type exampleWalk struct {
		Head  string
		Child *exampleWalk
		Tail  string
	}
	trnew := func(pnt *exampleWalk, i int) *exampleWalk {
		tr := &exampleWalk{
			Head:  fmt.Sprintf("tail #%d", i),
			Child: pnt,
			Tail:  fmt.Sprintf("tail #%d", i),
		}
		return tr
	}
	tr := trnew(nil, 4)
	for i := 3; i > 0; i-- {
		tr = trnew(tr, i)
	}

	w := iter.NewWalker(&iter.Iter{ChanRecv: true})
	err := w.Walk(tr, func(el iter.Pair) error {
		pad := strings.Repeat("  ", el.Depth())
		k, v := el.Pair()
		if sf, ok := k.(reflect.StructField); ok {
			fmt.Printf("%v%v -> %v\n", pad, sf.Name, v)
		} else {
			fmt.Printf("%v%v -> %v\n", pad, k, v)
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// Head -> tail #1
	//     Head -> tail #2
	//       Head -> tail #3
	//         Head -> tail #4
	//         Child -> <nil>
	//         Tail -> tail #4
	//       Tail -> tail #3
	//     Tail -> tail #2
	//   Tail -> tail #1

}

package iter

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type invalidIter struct {
	Iterator
	zeroValue       reflect.Value
	zeroStructField reflect.StructField
}

func (it invalidIter) IterMap(val reflect.Value, f func(key, val reflect.Value) error) (err error) {
	return f(it.zeroValue, it.zeroValue)
}

func (it invalidIter) IterSlice(val reflect.Value, f func(idx int, val reflect.Value) error) (err error) {
	return f(0, it.zeroValue)
}

func (it invalidIter) IterStruct(val reflect.Value, f func(field reflect.StructField, val reflect.Value) error) (err error) {
	return f(it.zeroStructField, it.zeroValue)
}

func (it invalidIter) IterChan(val reflect.Value, f func(seq int, recv reflect.Value) error) (err error) {
	return f(0, it.zeroValue)
}

func tchkstr(t testing.TB, err error, errStr string) error {
	if err != nil {
		if len(errStr) == 0 {
			return err
		}
		if !strings.Contains(err.Error(), errStr) {
			return fmt.Errorf("error did not match:\n  exp: %v\n  got: %v", errStr, err)
		}
	} else if len(errStr) > 0 {
		return errors.New("expected non-nil err")
	}
	return nil
}

func kindLength(k reflect.Kind) bool {
	return reflect.Array == k || reflect.Chan == k ||
		reflect.Map == k || reflect.Slice == k ||
		reflect.String == k
}

func TestIteratorInterface(t *testing.T) {
	var _ Iterator = Iter{}
	var _ Iterator = (*Iter)(nil)
	var _ Iterator = recoverIter{}
	var _ Iterator = (*recoverIter)(nil)
}

func TestIterMap(t *testing.T) {
	type testIterMapResult struct{ key, val reflect.Value }
	type testIterMap struct {
		errStr string
		it     Iterator
		give   interface{}
		res    []testIterMapResult
		f      func(key, val reflect.Value) error
	}
	tests := make(map[string]*testIterMap)
	tnew := func(it Iterator, give interface{}, errStr string) *testIterMap {
		tc := &testIterMap{it: it, give: give, errStr: errStr}
		tc.f = func(key, val reflect.Value) error {
			tc.res = append(tc.res, testIterMapResult{key, val})
			return nil
		}
		return tc
	}
	tadderr := func(name string, give interface{}, errStr string) {
		tests[name+"UsingIter"] = tnew(&Iter{}, give, errStr)
		tests[name+"UsingRecover"] = tnew(&recoverIter{&Iter{}}, give, errStr)
	}
	tadd := func(name string, give interface{}) {
		tadderr(name, give, ``)
	}

	tadd("StringInt", map[string]int{"a": 1, "b": 2, "c": 3})
	tadd("IntString", map[int]string{1: "a", 2: "b", 3: "c"})
	tadd("StringString", map[string]string{"a": "1", "b": "2", "c": "3"})
	tadd("IntInt", map[int]int{1: 1, 2: 2, 3: 3})
	tadd("NestedMap", map[string]map[int]int{"m1": {1: 2, 3: 4}, "m2": {3: 4, 5: 6}})
	tadd("NestedSlice", map[string][]int{"m1": {1: 2, 3: 4}, "m2": {3: 4, 5: 6}})
	tadd("Empty", make(map[int]int))
	tadd("NilMap", (map[int]int)(nil))

	tadderr("Nil", nil, "expected map kind, not invalid")
	tadderr("EmptyString", "", "expected map kind, not string")
	tadderr("WrongIterable", []string{"1", "2"}, "expected map kind, not slice")

	for name, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%v", name), func(t *testing.T) {
			val := reflect.ValueOf(tc.give)
			valKind := val.Kind()
			valLen := 0
			if kindLength(valKind) {
				valLen = val.Len()
			} else if valKind == reflect.Struct {
				valLen = val.NumField()
			}

			if err := tchkstr(t, tc.it.IterMap(val, tc.f), tc.errStr); err != nil {
				t.Fatal(err)
			}
			if len(tc.errStr) > 0 {
				return
			}

			if len(tc.res) != valLen {
				t.Fatalf("expected %d elements, got %d", valLen, len(tc.res))
			}
			for _, r := range tc.res {
				if !r.key.IsValid() {
					t.Fatalf("invalid key in result: key(%v) val(%v)", r.key, r.val)
				}
				if !r.val.IsValid() {
					t.Fatalf("invalid val in result: key(%v) val(%v)", r.key, r.val)
				}
				if !r.val.CanInterface() {
					t.Fatalf("cant interface val in result: key(%v) val(%v)", r.key, r.val)
				}

				expVal := val.MapIndex(r.key)
				if !expVal.IsValid() {
					t.Fatalf("could not find validation key for result: key(%v) val(%v)", r.key, r.val)
				}
				if !expVal.CanInterface() {
					t.Fatalf("invalid validation value (%v) for result: key(%v) val(%v)", expVal, r.key, r.val)
				}

				exp := expVal.Interface()
				got := r.val.Interface()
				if !reflect.DeepEqual(exp, got) {
					t.Errorf("DeepEqual failed for key(%v):\n  exp: %#v\n  got: %#v", r.key, exp, got)
				}
			}

			if len(tc.res) > 0 {
				expErr := errors.New("propagate error")
				errf := func(key, val reflect.Value) error {
					return expErr
				}
				if err := tc.it.IterMap(val, errf); err != expErr {
					t.Error("error did not propagate")
				}
			}
		})
	}
}

func TestIterSlice(t *testing.T) {
	type testIterSliceResult struct {
		idx int
		val reflect.Value
	}
	type testIterSlice struct {
		errStr string
		it     Iterator
		give   interface{}
		res    []testIterSliceResult
		f      func(idx int, val reflect.Value) error
	}
	tests := make(map[string]*testIterSlice)
	tnew := func(it Iterator, give interface{}, errStr string) *testIterSlice {
		tc := &testIterSlice{it: it, give: give, errStr: errStr}
		tc.f = func(idx int, val reflect.Value) error {
			tc.res = append(tc.res, testIterSliceResult{idx, val})
			return nil
		}
		return tc
	}
	tadderr := func(name string, give interface{}, errStr string) {
		tests[name+"UsingIter"] = tnew(&Iter{}, give, errStr)
		tests[name+"UsingNewIter"] = tnew(NewIter(), give, errStr)
		tests[name+"UsingRecover"] = tnew(&recoverIter{&Iter{}}, give, errStr)
		tests[name+"UsingNewRecover"] = tnew(NewRecoverIter(NewIter()), give, errStr)
	}
	tadd := func(name string, give interface{}) {
		tadderr(name, give, ``)
	}

	tadd("Ints", []int{1, 2, 3})
	tadd("IntsArr", [3]int{1, 2, 3})
	tadd("Strings", []string{"a", "b", "c"})
	tadd("StringsArr", [3]string{"a", "b", "c"})
	tadd("StringsJagged", [][]string{{"a", "b", "c"}, {"d", "e", "f"}})
	tadd("StringsArrJagged", [2][3]string{{"a", "b", "c"}, {"d", "e", "f"}})

	tadd("Empty", []int{})
	tadd("EmptyArr", [0]int{})
	tadd("EmptyArr10", [10]int{})
	tadd("NilSlice", ([]int)(nil))

	tadderr("Nil", nil, "expected array or slice kind, not invalid")
	tadderr("EmptyString", "", "expected array or slice kind, not string")
	tadderr("WrongIterable", map[int]string{1: "1", 2: "2"}, "expected array or slice kind, not map")

	for name, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%v", name), func(t *testing.T) {
			val := reflect.ValueOf(tc.give)
			valKind := val.Kind()
			valLen := 0
			if kindLength(valKind) {
				valLen = val.Len()
			} else if valKind == reflect.Struct {
				valLen = val.NumField()
			}

			if err := tchkstr(t, tc.it.IterSlice(val, tc.f), tc.errStr); err != nil {
				t.Fatal(err)
			}
			if len(tc.errStr) > 0 {
				return
			}

			if len(tc.res) != valLen {
				t.Fatalf("expected %d elements, got %d", valLen, len(tc.res))
			}
			for _, r := range tc.res {
				if !r.val.IsValid() {
					t.Fatalf("invalid val in result: idx(%v) val(%v)", r.idx, r.val)
				}
				if !r.val.CanInterface() {
					t.Fatalf("cant interface val in result: idx(%v) val(%v)", r.idx, r.val)
				}

				expVal := val.Index(r.idx)
				if !expVal.IsValid() {
					t.Fatalf("could not find validation element for result: idx(%v) val(%v)", r.idx, r.val)
				}
				if !expVal.CanInterface() {
					t.Fatalf("invalid validation value (%v) for result: idx(%v) val(%v)", expVal, r.idx, r.val)
				}

				exp := expVal.Interface()
				got := r.val.Interface()
				if !reflect.DeepEqual(exp, got) {
					t.Errorf("DeepEqual failed for index(%v):\n  exp: %#v\n  got: %#v", r.idx, exp, got)
				}
			}

			if len(tc.res) > 0 {
				expErr := errors.New("propagate error")
				errf := func(idx int, val reflect.Value) error {
					return expErr
				}
				if err := tc.it.IterSlice(val, errf); err != expErr {
					t.Error("error did not propagate")
				}
			}
		})
	}
}

func TestIterStruct(t *testing.T) {
	type testIterStructResult struct {
		field reflect.StructField
		val   reflect.Value
	}
	type testIterStructOpts struct {
		ExcludeAnonymous  bool
		ExcludeUnexported bool
	}
	type testIterStruct struct {
		opts   testIterStructOpts
		errStr string
		it     Iterator
		give   interface{}
		res    []testIterStructResult
		f      func(field reflect.StructField, val reflect.Value) error
	}
	tests := make(map[string]*testIterStruct)
	tnew := func(it Iterator, give interface{}, errStr string, opts testIterStructOpts) *testIterStruct {
		tc := &testIterStruct{it: it, give: give, errStr: errStr, opts: opts}
		tc.f = func(field reflect.StructField, val reflect.Value) error {
			tc.res = append(tc.res, testIterStructResult{field, val})
			return nil
		}
		return tc
	}

	var it *Iter
	tadderr := func(name string, give interface{}, errStr string) {
		opts := testIterStructOpts{it.ExcludeAnonymous, it.ExcludeUnexported}
		tests[name+"UsingIter"] = tnew(it, give, errStr, opts)
		tests[name+"UsingRecover"] = tnew(&recoverIter{it}, give, errStr, opts)
	}
	tadd := func(name string, give interface{}) {
		tadderr(name, give, ``)
	}
	addtests := func() {
		tadd("Ints", struct{ A, B, C int }{1, 2, 3})
		tadd("Mixed", struct {
			A, B int
			C    string
		}{1, 2, "A"})
		tadd("IntsField", struct {
			A int
			B string
			C []int
		}{1, "A", []int{2, 3}})
		tadd("MapField", struct {
			A int
			B string
			C map[string]int
		}{1, "A", map[string]int{"B": 2, "C": 3}})

		tadd("HeadUnexported", struct{ a, B, C int }{1, 2, 3})
		tadd("MidUnexported", struct{ A, b, C int }{1, 2, 3})
		tadd("TailUnexported", struct{ A, B, c int }{1, 2, 3})
		tadd("AllUnexported", struct{ a, b, c int }{1, 2, 3})

		var anonymousTest bytes.Buffer
		tadd("Anonymous", reflect.ValueOf(anonymousTest))

		tadd("Empty", struct{}{})
		tadderr("EmptyString", "", "expected struct kind, not string")
		tadderr("WrongIterable", map[int]string{1: "1", 2: "2"}, "expected struct kind, not map")
	}

	it = &Iter{ExcludeAnonymous: false, ExcludeUnexported: false}
	addtests()
	it = &Iter{ExcludeAnonymous: true, ExcludeUnexported: false}
	addtests()
	it = &Iter{ExcludeAnonymous: false, ExcludeUnexported: true}
	addtests()
	it = &Iter{ExcludeAnonymous: true, ExcludeUnexported: true}
	addtests()

	for name, tc := range tests {
		tc := tc
		s := fmt.Sprintf("%v/ExcludeAnonymous[%v]/ExcludeUnexported[%v]",
			name, tc.opts.ExcludeAnonymous, tc.opts.ExcludeUnexported)
		t.Run(s, func(t *testing.T) {
			val := reflect.ValueOf(tc.give)
			valKind := val.Kind()
			valLen := 0
			if kindLength(valKind) {
				valLen = val.Len()
			} else if valKind == reflect.Struct {
				typ := val.Type()
				for i := 0; i < val.NumField(); i++ {
					field := typ.Field(i)
					if field.Anonymous && tc.opts.ExcludeAnonymous {
						continue
					}
					if len(field.PkgPath) > 0 && tc.opts.ExcludeUnexported {
						continue
					}
					valLen++
				}
			}

			if err := tchkstr(t, tc.it.IterStruct(val, tc.f), tc.errStr); err != nil {
				t.Fatal(err)
			}
			if len(tc.errStr) > 0 {
				return
			}

			if len(tc.res) != valLen {
				t.Fatalf("expected %d elements, got %d", valLen, len(tc.res))
			}
			for _, r := range tc.res {
				if !r.val.IsValid() {
					t.Fatalf("invalid val in result: field(%v) val(%v)", r.field, r.val)
				}
				if !r.val.CanInterface() {
					// It's qualified by package path if unexported
					if len(r.field.PkgPath) > 0 && !tc.opts.ExcludeUnexported {
						t.Logf("qualified by pkg path %v as expected by ExcludeUnexported = false", r.field.PkgPath)
						continue
					}

					t.Fatalf("cant interface val in result: field(%v) val(%v)", r.field, r.val)
				}

				expVal := val.FieldByIndex(r.field.Index)
				if !expVal.IsValid() {
					t.Fatalf("could not find validation key for result: field(%v) val(%v)", r.field, r.val)
				}
				if !expVal.CanInterface() {
					t.Fatalf("invalid validation value (%v) for result: field(%v) val(%v)", expVal, r.field, r.val)
				}

				exp := expVal.Interface()
				got := r.val.Interface()
				if !reflect.DeepEqual(exp, got) {
					t.Errorf("DeepEqual failed for field(%v):\n  exp: %#v\n  got: %#v", r.field.Name, exp, got)
				}
			}

			if len(tc.res) > 0 {
				expErr := errors.New("propagate error")
				errf := func(field reflect.StructField, val reflect.Value) error {
					return expErr
				}
				if err := tc.it.IterStruct(val, errf); err != expErr {
					t.Error("error did not propagate")
				}
			}
		})
	}
}

func TestIterChan(t *testing.T) {
	type testIterChanResult struct {
		val reflect.Value
	}
	type testIterChanOpts struct {
		ChanBlock bool
		ChanRecv  bool
	}
	type testIterChan struct {
		opts   testIterChanOpts
		errStr string
		it     Iterator
		chf    func() interface{}
		give   interface{}
		res    []testIterChanResult
		gof    func(ch, give interface{})
		f      func(seq int, recv reflect.Value) error
	}
	tests := make(map[string]*testIterChan)
	tnew := func(it Iterator, chf func() interface{}, give interface{}, gof func(ch, give interface{}), errStr string, opts testIterChanOpts) *testIterChan {
		tc := &testIterChan{it: it, chf: chf, give: give, errStr: errStr, gof: gof, opts: opts}
		tc.f = func(seq int, recv reflect.Value) error {
			tc.res = append(tc.res, testIterChanResult{recv})
			return nil
		}
		return tc
	}

	var it *Iter
	tadderr := func(name string, chf func() interface{}, give interface{}, gof func(ch, give interface{}), errStr string) {
		opts := testIterChanOpts{it.ChanBlock, it.ChanRecv}
		tests[name+"UsingIter"] = tnew(it, chf, give, gof, errStr, opts)
		tests[name+"UsingRecover"] = tnew(&recoverIter{it}, chf, give, gof, errStr, opts)
	}
	tadd := func(name string, chf func() interface{}, give interface{}, gof func(ch, give interface{})) {
		tadderr(name, chf, give, gof, ``)
	}

	it = &Iter{ChanBlock: false, ChanRecv: true}
	{
		give := []string{"A", "B", "C"}
		chf := func() interface{} {
			ch := make(chan string, 3)
			for _, v := range give {
				ch <- v
			}
			return ch
		}
		tadd("Strings", chf, give, func(ch, give interface{}) {})
	}
	{
		give := []int{1, 2, 3}
		chf := func() interface{} {
			ch := make(chan int, 3)
			for _, v := range give {
				ch <- v
			}
			return ch
		}
		tadd("Ints", chf, give, func(ch, give interface{}) {})
	}
	{
		type StructPayload struct{ A, B int }
		give := []StructPayload{{1, 2}, {3, 4}, {5, 6}}
		chf := func() interface{} {
			ch := make(chan StructPayload, 3)
			for _, v := range give {
				ch <- v
			}
			return ch
		}
		tadd("Structs", chf, give, func(ch, give interface{}) {})
	}

	it = &Iter{ChanBlock: true, ChanRecv: true}
	{
		type StructPayload struct{ A, B int }
		give := []StructPayload{{1, 2}, {3, 4}, {5, 6}}
		chf := func() interface{} {
			return make(chan StructPayload)
		}
		gof := func(inch, give interface{}) {
			ch := inch.(chan StructPayload)
			for _, v := range give.([]StructPayload) {
				ch <- v
			}
			close(ch)
		}
		tadd("StructsBlock", chf, give, gof)
	}

	it = &Iter{ChanBlock: false, ChanRecv: true}
	{
		chf := func() interface{} {
			return ""
		}
		tadderr("NotChan", chf, nil, func(ch, give interface{}) {},
			`expected chan kind, not string`)
	}
	{
		chf := func() interface{} {
			return []string{""}
		}
		tadderr("ChanIterable", chf, nil, func(ch, give interface{}) {},
			`expected chan kind, not slice`)
	}
	{
		chf := func() interface{} {
			return nil
		}
		tadderr("NilChan", chf, nil, func(ch, give interface{}) {},
			`expected chan kind, not invalid`)
	}
	{
		chf := func() interface{} {
			return func() {}
		}
		tadderr("NilChan", chf, nil, func(ch, give interface{}) {},
			`expected chan kind, not func`)
	}

	t.Run("ChanReceive[false]", func(t *testing.T) {
		it = &Iter{}
		ch := make(chan string)
		it.IterChan(reflect.ValueOf(ch), func(seq int, val reflect.Value) error {
			return nil
		})
	})

	for name, tc := range tests {
		tc := tc
		s := fmt.Sprintf("%v/ChanBlock[%v]/ChanRecv[%v]",
			name, tc.opts.ChanBlock, tc.opts.ChanRecv)

		t.Run(s, func(t *testing.T) {
			val := reflect.ValueOf(tc.give)
			valKind := val.Kind()
			valLen := 0

			if kindLength(valKind) {
				valLen = val.Len()
			} else if valKind == reflect.Struct {
				valLen++
			}

			ch := tc.chf()
			chv := reflect.ValueOf(ch)
			go tc.gof(ch, tc.give)

			if err := tchkstr(t, tc.it.IterChan(chv, tc.f), tc.errStr); err != nil {
				t.Fatal(err)
			}
			if len(tc.errStr) > 0 {
				return
			}

			if len(tc.res) != valLen {
				t.Fatalf("expected %d elements, got %d", valLen, len(tc.res))
			}
			for i, r := range tc.res {
				if !r.val.IsValid() {
					t.Fatalf("invalid val in %T: val(%v)", ch, r.val)
				}
				if !r.val.CanInterface() {
					t.Fatalf("cant interface val in %T: val(%v)", ch, r.val)
				}

				expVal := val.Index(i)
				if !expVal.IsValid() {
					t.Fatalf("could not find validation element in %T:  val(%v)", ch, r.val)
				}
				if !expVal.CanInterface() {
					t.Fatalf("invalid validation value (%v) in %T: val(%v)", expVal, ch, r.val)
				}

				exp := expVal.Interface()
				got := r.val.Interface()
				if !reflect.DeepEqual(exp, got) {
					t.Errorf("DeepEqual failed in %T:\n  exp: %#v\n  got: %#v", ch, exp, got)
				}
			}

			ch = tc.chf()
			chv = reflect.ValueOf(ch)
			go tc.gof(ch, tc.give)

			if len(tc.res) > 0 {
				expErr := errors.New("propagate error")
				errf := func(seq int, val reflect.Value) error {
					if seq == valLen-1 { // drain channel
						return expErr
					}
					return nil
				}
				if err := tc.it.IterChan(chv, errf); err != expErr {
					t.Error("error did not propagate")
				}
			}
		})
	}
}

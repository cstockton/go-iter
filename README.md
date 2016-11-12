# Go Package: iter

  <a href="https://godoc.org/github.com/cstockton/go-iter"><img src="https://img.shields.io/badge/%20docs-reference-5272B4.svg?style=flat-square"></a> [![Go Report Card](https://goreportcard.com/badge/github.com/cstockton/go-iter)](https://goreportcard.com/report/github.com/cstockton/go-iter)

  > Get:
  > ```bash
  > go get -u github.com/cstockton/go-iter
  > ```
  >
  > Example:
  > ```Go
  > func Example() {
  > 	v := []interface{}{"a", "b", []string{"c", "d"}}
  > 	err := iter.Walk(v, func(el iter.Pair) error {
  > 		// check for errors
  > 		if err := el.Err(); err != nil {
  > 			return err
  > 		}
  >
  > 		// Halt iteration by returning an error.
  > 		if el.Depth() > 1 {
  > 			return errors.New("Stopping this walk.")
  > 		}
  >
  > 		fmt.Println(el)
  > 		return nil
  > 	})
  > 	if err == nil {
  > 		log.Fatal(err)
  > 	}
  >
  > 	// Output:
  > 	// Pair{(int) 0 => a (string)}
  > 	// Pair{(int) 1 => b (string)}
  > }
  > ```


## About

Package iter provides primitives for walking arbitrary data structures.


## Bugs and Patches

  Feel free to report bugs and submit pull requests.

  * bugs:
    <https://github.com/cstockton/go-iter/issues>
  * patches:
    <https://github.com/cstockton/go-iter/pulls>

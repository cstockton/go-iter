package iter

import (
	"fmt"
	"reflect"
	"testing"
)

func BenchmarkIteration(b *testing.B) {
	for x := uint64(4); x < 18; x += 4 {
		benchCount := 1 << (x - 1)

		b.Run(fmt.Sprintf("Slices/%d", benchCount), func(b *testing.B) {
			benchStrings := make([]string, benchCount)
			for y := 0; y < benchCount; y++ {
				benchStrings[y] = fmt.Sprintf("bench string %v", y)
			}
			benchVal := reflect.ValueOf(benchStrings)
			benchVals := make([]reflect.Value, len(benchStrings))
			for i, v := range benchStrings {
				benchVals[i] = reflect.ValueOf(v)
			}
			b.ResetTimer()

			b.Run("Range", func(b *testing.B) {
				f := func(idx int, v reflect.Value) error {
					return nil
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for z, v := range benchVals {
						f(z, v)
					}
				}
			})
			b.Run("Iter", func(b *testing.B) {
				it := &Iter{}
				f := func(idx int, v reflect.Value) error {
					return nil
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					it.IterSlice(benchVal, f)
				}
			})
			b.Run("RecoverIter", func(b *testing.B) {
				it := recoverIter{&Iter{}}
				f := func(idx int, v reflect.Value) error {
					return nil
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					it.IterSlice(benchVal, f)
				}
			})
		})
		b.Run(fmt.Sprintf("Maps/%d", benchCount), func(b *testing.B) {
			benchMaps := make(map[int]string, benchCount)
			for y := 0; y < benchCount; y++ {
				benchMaps[y] = fmt.Sprintf("bench string %v", y)
			}
			benchVal := reflect.ValueOf(benchMaps)
			benchVals := make([]reflect.Value, len(benchMaps))
			for i, v := range benchMaps {
				benchVals[i] = reflect.ValueOf(v)
			}
			b.ResetTimer()

			b.Run("Range", func(b *testing.B) {
				f := func(k, v reflect.Value) error {
					return nil
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for z, v := range benchVals {
						f(benchVals[z], v)
					}
				}
			})
			b.Run("Iter", func(b *testing.B) {
				it := &Iter{}
				f := func(k, v reflect.Value) error {
					return nil
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					it.IterMap(benchVal, f)
				}
			})
			b.Run("RecoverIter", func(b *testing.B) {
				it := recoverIter{&Iter{}}
				f := func(idx int, v reflect.Value) error {
					return nil
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					it.IterSlice(benchVal, f)
				}
			})
		})
	}
}

package utils

import (
	"reflect"

	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer
}

// ValToIdx는 비트로 표현된 값을 배열로 만들어 준다.
func ValToIdx[T interface{ Number }](v T) []int {
	var ret []int
	i := 0
	for v > 0 {
		if (v & (1 << i)) != 0 {
			ret = append(ret, i+1)
			v = v ^ (1 << i)
		}
		i++
	}
	return ret
}

// src, dst는 포인터 타입이어야 함.
func DeepCopy(src, dst interface{}) {
	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()
	DeepCopyValue(srcVal, dstVal)
}

func DeepCopyValue(src, dst reflect.Value) {
	switch src.Kind() {
	case reflect.Pointer:
		if !src.IsNil() {
			dst.Set(reflect.New(src.Elem().Type()))
			DeepCopyValue(src.Elem(), dst.Elem())
		}
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			DeepCopyValue(src.Field(i), dst.Field(i))
		}
	case reflect.Slice:
		if !src.IsNil() {
			dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
			for i := 0; i < src.Len(); i++ {
				DeepCopyValue(src.Index(i), dst.Index(i))
			}
		}
	default:
		dst.Set(src)
	}
}

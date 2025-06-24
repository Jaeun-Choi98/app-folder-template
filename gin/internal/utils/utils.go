package utils

import (
	"fmt"
	"reflect"
	"time"

	"golang.org/x/exp/constraints"
)

var (
	LocalKorea = NewKoreaLocation()
)

func NewKoreaLocation() *time.Location {
	local, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return nil
	}
	return local
}

type Number interface {
	constraints.Integer
}

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

func IdxToVal[T Number](idxs []int) T {
	var result T = 0

	for _, idx := range idxs {
		if idx > 0 {
			result |= T(1 << (idx - 1))
		}
	}

	return result
}

/**
* 깊은 복사
* @param src interface{} - 복사할 구조체 (포인터)
* @param dst는 interface{} - 목적지 구조체 (포인터)
**/
func DeepCopy(src, dst interface{}) {
	s := reflect.ValueOf(src)
	d := reflect.ValueOf(dst)

	// 포인터가 아니면 수정 불가
	if s.Kind() != reflect.Ptr || d.Kind() != reflect.Ptr {
		return
	}

	DeepCopyValue(s.Elem(), d.Elem())
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

/**
* 구조체 필드를 순차적으로 수정하는 표준 함수
* @param s interface{} - 수정할 구조체 (포인터)
* @param startIdx int - 시작 필드 인덱스
* @param endIdx int - 종료 필드 인덱스 (-1이면 마지막까지)
* @param params ...interface{} - 순차적으로 적용할 값들
**/
func ModifyStructFields(s interface{}, startIdx int, endIdx int, params ...interface{}) error {
	v := reflect.ValueOf(s)

	// 포인터가 아니면 수정 불가
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("utils err: modifying struct fields line 17: not a pointer")
	}

	// 실제 값 가져오기
	v = v.Elem()
	t := v.Type()

	totalFields := v.NumField()

	// 인덱스 검증 및 조정
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx < 0 || endIdx >= totalFields {
		endIdx = totalFields - 1
	}
	if startIdx > endIdx {
		return fmt.Errorf("utils err: modifying struct fields line 33: start index %d > end index %d", startIdx, endIdx)
	}

	paramIdx := 0

	for i := startIdx; i <= endIdx && i < totalFields; i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 수정 가능한지 확인
		if !field.CanSet() {
			continue
		}

		// 파라미터가 있으면 사용, 없으면 스킵
		if paramIdx >= len(params) {
			break
		}
		newValue := params[paramIdx]

		// 타입이 정확히 일치하는지 확인 후 설정
		if err := setFieldValue(field, newValue, i, fieldType.Name); err != nil {
			return err
		}
		paramIdx++
	}
	return nil
}

func setFieldValue(field reflect.Value, newValue interface{}, fieldIndex int, fieldName string) error {
	if newValue == nil {
		return fmt.Errorf("utils err: modifying struct fields line 72: field[%d] %s cannot set nil value", fieldIndex, fieldName)
	}

	newVal := reflect.ValueOf(newValue)
	fieldType := field.Type()

	if newVal.Type() != fieldType {
		return fmt.Errorf("utils err: modifying struct fields line 80: field[%d] %s type mismatch: expected %s, got %s",
			fieldIndex, fieldName, fieldType, newVal.Type())
	}

	field.Set(newVal)
	return nil
}

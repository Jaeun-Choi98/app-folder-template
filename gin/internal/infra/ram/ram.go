package ram

import (
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/object"
	"reflect"
	"sync"
)

type Ram struct {
	mu          sync.RWMutex
	dao         dbhandler.DBHandlerInterface
	SampleCache map[int]*object.Sample
}

func NewRam(dao dbhandler.DBHandlerInterface) (*Ram, error) {

	ram := &Ram{
		dao: dao,
	}

	if err := ram.LoadSampleCache(); err != nil {
		return nil, err
	}
	return ram, nil
}

// src, dst는 포인터 타입이어야 함.
func deepCopy(src, dst interface{}) {
	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()
	deepCopyValue(srcVal, dstVal)
}

func deepCopyValue(src, dst reflect.Value) {
	switch src.Kind() {
	case reflect.Pointer:
		if !src.IsNil() {
			dst.Set(reflect.New(src.Elem().Type()))
			deepCopyValue(src.Elem(), dst.Elem())
		}
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			deepCopyValue(src.Field(i), dst.Field(i))
		}
	case reflect.Slice:
		if !src.IsNil() {
			dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
			for i := 0; i < src.Len(); i++ {
				deepCopyValue(src.Index(i), dst.Index(i))
			}
		}
	default:
		dst.Set(src)
	}
}

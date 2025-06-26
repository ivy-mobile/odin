package xconv

import (
	"reflect"

	"github.com/ivy-mobile/odin/xutil/xreflect"
)

func Anys(data any) []any {
	if data == nil {
		return nil
	}

	switch rk, rv := xreflect.Value(data); rk {
	case reflect.Slice, reflect.Array:
		count := rv.Len()
		slice := make([]any, count)
		for i := 0; i < count; i++ {
			slice[i] = rv.Index(i).Interface()
		}
		return slice
	default:
		return nil
	}
}

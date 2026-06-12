package xconv

import (
	"reflect"

	"github.com/ivy-mobile/odin/encoding/json"
)

//nolint:revive // 保持既有导出 API 兼容。
func Json(data any) string {
	isJSON := func(s string) bool {
		l := len(s)
		return l >= 2 && ((s[0] == '{' && s[l-1] == '}') || (s[0] == '[' && s[l-1] == ']'))
	}

	switch v := data.(type) {
	case string:
		if isJSON(v) {
			return v
		}
	case *string:
		if isJSON(*v) {
			return *v
		}
	case []byte:
		if s := BytesToString(v); isJSON(s) {
			return s
		}
	case *[]byte:
		if s := BytesToString(*v); isJSON(s) {
			return s
		}
	default:
		var (
			rv   = reflect.ValueOf(data)
			kind = rv.Kind()
		)

		for kind == reflect.Pointer {
			rv = rv.Elem()
			kind = rv.Kind()
		}

		switch kind {
		case reflect.String:
			if s := rv.String(); isJSON(s) {
				return s
			}
		case reflect.Map, reflect.Array, reflect.Slice, reflect.Struct:
			if b, err := json.Marshal(v); err == nil {
				return BytesToString(b)
			}
		}
	}

	return ""
}

package xconv

import "reflect"

func Int16(data any) int16 {
	return int16(Int64(data))
}

//nolint:gocyclo // 覆盖多种切片/数组输入，保持既有显式分支语义。
func Int16s(data any) (slice []int16) {
	if data == nil {
		return
	}

	switch v := data.(type) {
	case []int:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]int:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []int8:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]int8:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []int16:
		return v
	case *[]int16:
		return *v
	case []int32:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]int32:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []int64:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]int64:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []uint:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]uint:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []uint8:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]uint8:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []uint16:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]uint16:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []uint32:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]uint32:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []uint64:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]uint64:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []float32:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]float32:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []float64:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]float64:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []complex64:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]complex64:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []complex128:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]complex128:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []string:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]string:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []bool:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]bool:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case []any:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[]any:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
		}
	case [][]byte:
		slice = make([]int16, len(v))
		for i := range v {
			slice[i] = Int16(v[i])
		}
	case *[][]byte:
		slice = make([]int16, len(*v))
		for i := range *v {
			slice[i] = Int16((*v)[i])
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
		case reflect.Slice, reflect.Array:
			count := rv.Len()
			slice = make([]int16, count)
			for i := 0; i < count; i++ {
				slice[i] = Int16(rv.Index(i).Interface())
			}
		}
	}

	return
}

func Int16Pointer(data any) *int16 {
	v := Int16(data)
	return &v
}

func Int16sPointer(data any) *[]int16 {
	v := Int16s(data)
	return &v
}

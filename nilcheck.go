package coroutine

import "reflect"

func IsNilValue(i interface{}) bool {
	if i == nil {
		return true
	}

	// 如果是带类型的nil，判断其值是否是nil
	if reflect.ValueOf(i).IsNil() {
		return true
	}
	return false
}

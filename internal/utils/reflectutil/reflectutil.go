// Package reflectutil Функции рефлексии
package reflectutil

import (
	"reflect"
	"runtime"
)

func GetFunctionName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

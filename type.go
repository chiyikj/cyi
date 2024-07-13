package cyi

import "reflect"

func isType(typeName reflect.Type) bool {
	if typeName.Kind() == reflect.Bool || (typeName.Kind() == reflect.Ptr && typeName.Elem().Kind() == reflect.Bool) {
		return true
	} else if typeName.Kind() == reflect.String || (typeName.Kind() == reflect.Ptr && typeName.Elem().Kind() == reflect.String) {
		return true
	} else if typeName.Kind() == reflect.Slice || (typeName.Kind() == reflect.Ptr && typeName.Elem().Kind() == reflect.Slice) {
		return true
	} else if typeName.Kind() == reflect.Struct || (typeName.Kind() == reflect.Ptr && typeName.Elem().Kind() == reflect.Struct) {
		return true
	} else if typeName.Kind() == reflect.Float64 || (typeName.Kind() == reflect.Ptr && typeName.Elem().Kind() == reflect.Float64) {
		return true
	} else if typeName.Kind() == reflect.Int || (typeName.Kind() == reflect.Ptr && typeName.Elem().Kind() == reflect.Int) {
		return true
	} else {
		return false
	}
}

func isComplexType(typeName reflect.Type) bool {
	if typeName.Kind() == reflect.Slice || (typeName.Kind() == reflect.Ptr && typeName.Elem().Kind() == reflect.Slice) {
		return true
	} else if typeName.Kind() == reflect.Struct || (typeName.Kind() == reflect.Ptr && typeName.Elem().Kind() == reflect.Struct) {
		return true
	} else {
		return false
	}
}

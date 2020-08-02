package mongoclient

import (
	"reflect"
	"strings"
)

func arrayContains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func structFieldValueByTag(s interface{}, tagKey, tagValue string) (reflect.Value, bool) {
	rt := reflect.TypeOf(s)
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		v := field.Tag.Get(tagKey)
		arr := strings.Split(v, ",")
		if arrayContains(arr, tagValue) {
			return reflect.ValueOf(s).FieldByIndex(field.Index), true
		}
	}
	return reflect.Value{}, false
}

func isStruct(i interface{}) bool {
	if reflect.TypeOf(i).Kind() == reflect.Struct {
		return true
	}

	return false
}

func isPointerOfStruct(i interface{}) bool {
	typeOf := reflect.TypeOf(i)

	if typeOf.Kind() == reflect.Ptr {
		return typeOf.Elem().Kind() == reflect.Struct
	}

	return false
}

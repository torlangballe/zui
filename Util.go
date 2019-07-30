package zgo

import (
	"errors"
	"reflect"
)

func checkSlice(s interface{}) error {
	if s == nil {
		return errors.New("slice is nil!")
	}

	slicePtrValue := reflect.ValueOf(s)
	// should be pointer
	if slicePtrValue.Type().Kind() != reflect.Ptr {
		return errors.New("should be slice pointer!")
	}

	sliceValue := slicePtrValue.Elem()
	// should be slice
	if sliceValue.Type().Kind() != reflect.Slice {
		return errors.New("should be slice pointer!")
	}

	return nil
}

func UtilRemoveAt(slice interface{}, index int) error {
	err := checkSlice(slice)
	if err != nil {
		return err
	}

	slicePtrValue := reflect.ValueOf(slice)
	sliceValue := slicePtrValue.Elem()
	if index < 0 || index >= sliceValue.Len() {
		return errors.New("index out of range!")
	}
	sliceValue.Set(reflect.AppendSlice(sliceValue.Slice(0, index), sliceValue.Slice(index+1, sliceValue.Len())))
	return nil
}

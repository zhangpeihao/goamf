// Copyright 2013, zhangpeihao All rights reserved.

package amf

import (
	"reflect"
	"strconv"
)

// An UnsupportedTypeError is returned by Marshal when attempting
// to encode an unsupported value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "unsupported type: " + e.Type.String()
}

type UnsupportedValueError struct {
	Value reflect.Value
	Str   string
}

func (e *UnsupportedValueError) Error() string {
	return "unsupported value: " + e.Str
}

type InvalidUTF8Error struct {
	S string
}

func (e *InvalidUTF8Error) Error() string {
	return "invalid UTF-8 in string: " + strconv.Quote(e.S)
}

type NameLengthOverflowError struct {
	Name string
}

func (e *NameLengthOverflowError) Error() string {
	return "Filed[" + e.Name + "] is too long to encode into AMF0"
}

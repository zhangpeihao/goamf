// Copyright 2013, zhangpeihao All rights reserved.

package amf

import (
	"reflect"
)

const (
	AMF0 = uint(0)
	AMF3 = uint(3)
)

const (
	AMF0_NUMBER_MARKER         = 0x00
	AMF0_BOOLEAN_MARKER        = 0x01
	AMF0_STRING_MARKER         = 0x02
	AMF0_OBJECT_MARKER         = 0x03
	AMF0_MOVIECLIP_MARKER      = 0x04
	AMF0_NULL_MARKER           = 0x05
	AMF0_UNDEFINED_MARKER      = 0x06
	AMF0_REFERENCE_MARKER      = 0x07
	AMF0_ECMA_ARRAY_MARKER     = 0x08
	AMF0_OBJECT_END_MARKER     = 0x09
	AMF0_STRICT_ARRAY_MARKER   = 0x0a
	AMF0_DATE_MARKER           = 0x0b
	AMF0_LONG_STRING_MARKER    = 0x0c
	AMF0_UNSUPPORTED_MARKER    = 0x0d
	AMF0_RECORDSET_MARKER      = 0x0e
	AMF0_XML_DOCUMENT_MARKER   = 0x0f
	AMF0_TYPED_OBJECT_MARKER   = 0x10
	AMF0_ACMPLUS_OBJECT_MARKER = 0x11
)

const (
	AMF0_MAX_STRING_LEN = 65535
)

const (
	AMF3_UNDEFINED_MARKER = 0x00
	AMF3_NULL_MARKER      = 0x01
	AMF3_FALSE_MARKER     = 0x02
	AMF3_TRUE_MARKER      = 0x03
	AMF3_INTEGER_MARKER   = 0x04
	AMF3_DOUBLE_MARKER    = 0x05
	AMF3_STRING_MARKER    = 0x06
	AMF3_XMLDOC_MARKER    = 0x07
	AMF3_DATE_MARKER      = 0x08
	AMF3_ARRAY_MARKER     = 0x09
	AMF3_OBJECT_MARKER    = 0x0a
	AMF3_XML_MARKER       = 0x0b
	AMF3_BYTEARRAY_MARKER = 0x0c
)

type Writer interface {
	Write(p []byte) (nn int, err error)
	WriteByte(c byte) error
}

type Reader interface {
	Read(p []byte) (n int, err error)
	ReadByte() (c byte, err error)
}

// Undefined Type
type Undefined struct{}

// Object Type
type Object map[string]interface{}

// stringValues is a slice of reflect.Value holding *reflect.StringValue.
// It implements the methods to sort by string.
type stringValues []reflect.Value

func (sv stringValues) Len() int           { return len(sv) }
func (sv stringValues) Swap(i, j int)      { sv[i], sv[j] = sv[j], sv[i] }
func (sv stringValues) Less(i, j int) bool { return sv.get(i) < sv.get(j) }
func (sv stringValues) get(i int) string   { return sv[i].String() }

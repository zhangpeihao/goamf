// Copyright 2013, zhangpeihao All rights reserved.

package amf

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
	"runtime"
	"sort"
	"strconv"
)

// MarshalAMF0 returns the AMF0 object Type encoding of v.
// Spec: http://download.macromedia.com/pub/labs/amf/amf0_spec_121207.pdf
//
// Marshal traverses the value v recursively.
//
// Marshal uses the following type-dependent default encodings:
//
// Boolean values encode as AMF0 Boolean Type.
//
// Floating point and Integer encode as AMF0 Number Type.
//
// String and []byte values encode as AMF0 String Type.
//
// Array and slice values encode as AMF0 Object Type
// nil slice encodes as the null AMF0 object.
//
// Struct values encode as AMF0 objects. Each exported struct field
// becomes a member of the object unless
//   - the field's tag is "-", or
//   - the field is empty and its tag specifies the "omitempty" option.
// The empty values are false, 0, any
// nil pointer or interface value, and any array, slice, map, or string of
// length zero. The object's default key string is the struct field name
// but can be specified in the struct field's tag value. The "amf0" object name in
// the struct field's tag value is the key name, followed by an optional comma
// and options. Examples:
//
//   // Field is ignored by this package.
//   Field int `amf:"-"`
//
//   // Field appears in AMF0 as name "myName".
//   Field int `amf:"myName"`
//
//   // Field appears in AMF0 as name "myName" and
//   // the field is omitted from the object if its value is empty,
//   // as defined above.
//   Field int `amf:"myName,omitempty"`
//
//   // Field appears in AMF0 as key "Field" (the default), but
//   // the field is skipped if empty.
//   // Note the leading comma.
//   Field int `amf:",omitempty"`
//
// The "string" option signals that a field is stored as string:
//
//    Int64String int64 `amf:",string"`
//
// Map values encode as AMF0 objects.
// The map's key type must be string; the object names are used directly
// as map keys.
//
// Pointer values encode as the value pointed to.
// A nil pointer encodes as the null Type.
//
// Interface values encode as the value contained in the interface.
// A nil interface value encodes as the null Type.
//
// Channel, complex, and function values cannot be encoded in AMF0.
// Attempting to encode such a value causes Marshal to return
// an UnsupportedTypeError.
//
// AMF0 cannot represent cyclic data structures and Marshal does not
// handle them.  Passing cyclic structures to Marshal will result in
// an infinite recursion.
//
func MarshalAMF0(v interface{}) ([]byte, error) {
	e := &encodeStateAMF0{}
	err := e.marshal(v)
	if err != nil {
		return nil, err
	}
	return e.Bytes(), nil
}

// An encodeStateAMF0 encodes AMF0 into a bytes.Buffer.
type encodeStateAMF0 struct {
	bytes.Buffer // accumulated output
	scratch      [64]byte
}

func (e *encodeStateAMF0) WriteMark(value byte) {
	err := e.WriteByte(value)
	if err != nil {
		e.error(err)
	}
}

func (e *encodeStateAMF0) WriteObjectEndMark() {
	_, err := e.Write([]byte{0x00, 0x00, AMF0_OBJECT_END_MARKER})
	if err != nil {
		e.error(err)
	}
}

func (e *encodeStateAMF0) WriteAMF0String(str string) {
	length := len(str)
	if length > AMF0_MAX_STRING_LEN {
		e.WriteMark(AMF0_LONG_STRING_MARKER)
		err := binary.Write(e, binary.BigEndian, uint32(length))
		if err != nil {
			e.error(err)
		}
	} else {
		e.WriteMark(AMF0_STRING_MARKER)
		err := binary.Write(e, binary.BigEndian, uint16(length))
		if err != nil {
			e.error(err)
		}
	}
	e.WriteString(str)
}

func (e *encodeStateAMF0) WriteAMF0Number(f float64) {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		e.error(&UnsupportedValueError{reflect.ValueOf(f), strconv.FormatFloat(f, 'g', -1, 64)})
	}
	e.WriteMark(AMF0_NUMBER_MARKER)
	err := binary.Write(e, binary.BigEndian, f)
	if err != nil {
		e.error(err)
	}
}

func (e *encodeStateAMF0) WriteAMF0Boolean(b bool) {
	e.WriteMark(AMF0_BOOLEAN_MARKER)

	if b {
		e.WriteByte(0x01)
	} else {
		e.WriteByte(0x00)
	}
}

func (e *encodeStateAMF0) marshal(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()
	e.reflectValue(reflect.ValueOf(v))
	return nil
}

func (e *encodeStateAMF0) error(err error) {
	panic(err)
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func (e *encodeStateAMF0) reflectValue(v reflect.Value) {
	if !v.IsValid() {
		e.WriteMark(AMF0_NULL_MARKER)
		return
	}

	switch v.Kind() {
	case reflect.Bool:
		e.WriteAMF0Boolean(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.WriteAMF0Number(float64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		e.WriteAMF0Number(float64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		e.WriteAMF0Number(v.Float())
	case reflect.String:
		e.WriteAMF0String(v.String())
	case reflect.Struct:
		e.WriteMark(AMF0_OBJECT_MARKER)
		for _, ef := range encodeFields(v.Type()) {
			fieldValue := v.Field(ef.i)
			if ef.omitEmpty && isEmptyValue(fieldValue) {
				continue
			}
			length := len(ef.tag)
			if length > AMF0_MAX_STRING_LEN {
				e.error(&NameLengthOverflowError{ef.tag})
			}
			err := binary.Write(e, binary.BigEndian, uint16(length))
			if err != nil {
				e.error(err)
			}
			e.WriteString(ef.tag)
			e.reflectValue(fieldValue)
		}
		e.WriteObjectEndMark()

	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			e.error(&UnsupportedTypeError{v.Type()})
		}
		if v.IsNil() {
			e.WriteMark(AMF0_NULL_MARKER)
			break
		}
		e.WriteMark(AMF0_OBJECT_MARKER)
		var sv stringValues = v.MapKeys()
		sort.Sort(sv)
		for _, k := range sv {
			keyName := k.String()
			length := len(keyName)
			if length > AMF0_MAX_STRING_LEN {
				e.error(&NameLengthOverflowError{keyName})
			}
			err := binary.Write(e, binary.BigEndian, uint16(length))
			if err != nil {
				e.error(err)
			}
			e.WriteString(keyName)
			e.reflectValue(v.MapIndex(k))
		}
		e.WriteObjectEndMark()

	case reflect.Slice:
		if v.IsNil() {
			e.WriteMark(AMF0_NULL_MARKER)
			break
		}
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// Byte slices get special treatment; arrays don't.
			e.WriteAMF0String(string(v.Bytes()))
			break
		}
		// Slices can be marshalled as nil, but otherwise are handled
		// as arrays.
		fallthrough
	case reflect.Array:
		// Array encoded as object
		e.WriteMark(AMF0_OBJECT_MARKER)
		n := v.Len()
		for i := 0; i < n; i++ {
			strindex := strconv.Itoa(i)
			err := binary.Write(e, binary.BigEndian, uint16(len(strindex)))
			if err != nil {
				e.error(err)
			}
			e.WriteString(strindex)
			e.reflectValue(v.Index(i))
		}
		e.WriteObjectEndMark()

	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			e.WriteMark(AMF0_NULL_MARKER)
			return
		}
		e.reflectValue(v.Elem())

	default:
		e.error(&UnsupportedTypeError{v.Type()})
	}
	return
}

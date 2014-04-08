// Copyright 2013, zhangpeihao All rights reserved.

package amf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
)

//-----------------------------------------------------------------------
// AMF0 Write functions

func WriteMarker(w Writer, mark byte) (n int, err error) {
	err = w.WriteByte(mark)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func WriteString(w Writer, str string) (n int, err error) {
	length := uint32(len(str))
	if length > 0xFFFF {
		err = w.WriteByte(AMF0_LONG_STRING_MARKER)
		if err != nil {
			return 0, err
		}
		err = WriteUTF8Long(w, str, length)
		length += 5
	} else {
		err = w.WriteByte(AMF0_STRING_MARKER)
		if err != nil {
			return 0, err
		}
		err = WriteUTF8(w, str, uint16(length))
		length += 3
	}
	if err != nil {
		return 1, err
	}
	return int(length), nil
}

func WriteUTF8(w Writer, s string, length uint16) error {
	err := binary.Write(w, binary.BigEndian, &length)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	return err
}

func WriteUTF8Long(w Writer, s string, length uint32) error {
	err := binary.Write(w, binary.BigEndian, &length)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	return err
}

func WriteDouble(w Writer, num float64) (n int, err error) {
	err = w.WriteByte(AMF0_NUMBER_MARKER)
	if err != nil {
		return 0, err
	}
	err = binary.Write(w, binary.BigEndian, num)
	if err != nil {
		return 1, err
	}
	return 9, nil
}

func WriteBoolean(w Writer, b bool) (n int, err error) {
	err = w.WriteByte(AMF0_BOOLEAN_MARKER)
	if err != nil {
		return 0, err
	}
	if b {
		err = w.WriteByte(1)
	} else {
		err = w.WriteByte(0)
	}
	if err != nil {
		return 1, err
	}
	return 2, nil
}

func WriteNull(w Writer) (n int, err error) {
	err = w.WriteByte(AMF0_NULL_MARKER)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func WriteUndefined(w Writer) (n int, err error) {
	err = w.WriteByte(AMF0_UNDEFINED_MARKER)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func WriteEcmaArray(w Writer, arr []interface{}) (n int, err error) {
	n, err = WriteMarker(w, AMF0_ECMA_ARRAY_MARKER)
	if err != nil {
		return
	}
	length := int32(len(arr))
	err = binary.Write(w, binary.BigEndian, &length)
	if err != nil {
		return
	}
	n += 4
	m := 0
	for index, value := range arr {
		m, err = WriteObjectName(w, fmt.Sprintf("%d", index))
		if err != nil {
			return
		}
		n += m
		m, err = WriteValue(w, value)
		if err != nil {
			return
		}
		n += m
	}
	m, err = WriteObjectEndMarker(w)
	return n + m, err
}

func WriteObjectMarker(w Writer) (n int, err error) {
	return WriteMarker(w, AMF0_OBJECT_MARKER)
}

func WriteObjectEndMarker(w Writer) (n int, err error) {
	return w.Write([]byte{0x00, 0x00, AMF0_OBJECT_END_MARKER})
}

func WriteObjectName(w Writer, name string) (n int, err error) {
	length := uint16(len(name))
	err = WriteUTF8(w, name, length)
	return int(length + 2), err
}

// Object's item order is uncertainty.
func WriteObject(w Writer, obj Object) (n int, err error) {
	n, err = WriteObjectMarker(w)
	if err != nil {
		return
	}
	m := 0
	for key, value := range obj {
		m, err = WriteObjectName(w, key)
		if err != nil {
			return
		}
		n += m
		m, err = WriteValue(w, value)
		if err != nil {
			return
		}
		n += m
	}
	m, err = WriteObjectEndMarker(w)
	return n + m, err
}

func WriteFields(w Writer, value reflect.Value) (int, error) {
	var n int

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		structField := value.Type().Field(i)

		if field.Kind() == reflect.Struct && structField.Anonymous {
			m, err := WriteFields(w, field)
			if err != nil {
				return n, err
			}
			n += m
		} else {
			m, err := WriteField(w, structField, field)
			if err != nil {
				return n, err
			}
			n += m
		}
	}

	return n, nil
}

func WriteField(w Writer, structField reflect.StructField, field reflect.Value) (int, error) {
	var n, m int

	name := structField.Tag.Get("amf")
	if name == "" {
		name = structField.Name
	}

	m, err := WriteObjectName(w, name)
	if err != nil {
		return n, err
	}
	n += m

	m, err = WriteValue(w, field.Interface())
	if err != nil {
		return n, err
	}
	n += m

	return n, nil
}

func WriteValue(w Writer, value interface{}) (n int, err error) {
	if value == nil {
		return WriteNull(w)
	}
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return WriteNull(w)
	}
	switch v.Kind() {
	case reflect.String:
		return WriteString(w, value.(string))
	case reflect.Bool:
		return WriteBoolean(w, v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return WriteDouble(w, float64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return WriteDouble(w, float64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		return WriteDouble(w, v.Float())
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		// Copy from WriteEcmaArray
		n, err = WriteMarker(w, AMF0_ECMA_ARRAY_MARKER)
		if err != nil {
			return
		}
		length := int32(v.Len())
		err = binary.Write(w, binary.BigEndian, &length)
		if err != nil {
			return
		}
		n += 4
		m := 0
		for index := int32(0); index < length; index++ {
			m, err = WriteObjectName(w, fmt.Sprintf("%d", index))
			if err != nil {
				return
			}
			n += m
			m, err = WriteValue(w, v.Index(int(index)).Interface())
			if err != nil {
				return
			}
			n += m
		}
		m, err = WriteObjectEndMarker(w)
		return n + m, err
	case reflect.Struct:
		n, err = WriteObjectMarker(w)
		if err != nil {
			return
		}

		m := 0
		m, err = WriteFields(w, v)
		if err != nil {
			return
		}
		n += m

		m, err = WriteObjectEndMarker(w)
		return n + m, err
	}
	if _, ok := value.(Undefined); ok {
		return WriteUndefined(w)
	} else if vt, ok := value.(Object); ok {
		return WriteObject(w, vt)
	} else if vt, ok := value.([]interface{}); ok {
		return WriteEcmaArray(w, vt)
	}
	return 0, errors.New("Unsupported type")
}

//-----------------------------------------------------------------------
// AMF0 Read functions
func ReadMarker(r Reader) (mark byte, err error) {
	return r.ReadByte()
}

func ReadString(r Reader) (str string, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return "", err
	}
	switch marker {
	case AMF0_STRING_MARKER:
		return ReadUTF8(r)
	case AMF0_LONG_STRING_MARKER:
		return ReadUTF8Long(r)
	}
	return "", errors.New("Type error")
}
func ReadUTF8(r Reader) (string, error) {
	var stringLength uint16
	err := binary.Read(r, binary.BigEndian, &stringLength)
	if err != nil {
		return "", err
	}
	if stringLength == 0 {
		return "", nil
	}
	data := make([]byte, stringLength)
	_, err = r.Read(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ReadUTF8Long(r Reader) (string, error) {
	var stringLength uint32
	err := binary.Read(r, binary.BigEndian, &stringLength)
	if err != nil {
		return "", err
	}
	if stringLength == 0 {
		return "", nil
	}
	data := make([]byte, stringLength)
	_, err = r.Read(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ReadDouble(r Reader) (num float64, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return 0, err
	}
	if marker != AMF0_NUMBER_MARKER {
		return 0, errors.New("Type error")
	}
	err = binary.Read(r, binary.BigEndian, &num)
	return
}

func ReadBoolean(r Reader) (b bool, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return false, err
	}
	if marker != AMF0_BOOLEAN_MARKER {
		return false, errors.New("Type error")
	}
	value, err := r.ReadByte()
	return bool(value != 0), nil
}

func ReadObjectName(r Reader) (name string, err error) {
	return ReadUTF8(r)
}

func ReadObject(r Reader) (obj Object, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return nil, err
	}
	if marker != AMF0_OBJECT_MARKER {
		return nil, errors.New("Type error")
	}
	return ReadObjectProperty(r)
}

func ReadObjectProperty(r Reader) (Object, error) {
	obj := make(Object)
	for {
		name, err := ReadUTF8(r)
		if err != nil {
			return nil, err
		}
		if name == "" {
			b, err := r.ReadByte()
			if err != nil {
				return nil, err
			}
			if b == AMF0_OBJECT_END_MARKER {
				break
			} else {
				return nil, errors.New("expect ObjectEndMarker here")
			}
		}
		if _, ok := obj[name]; ok {
			return nil, errors.New("object-property exists")
		}

		value, err := ReadValue(r)
		if err != nil {
			return nil, err
		}
		obj[name] = value
	}
	return obj, nil
}

func ReadValue(r Reader) (value interface{}, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return 0, err
	}
	switch marker {
	case AMF0_NUMBER_MARKER:
		var num float64
		err = binary.Read(r, binary.BigEndian, &num)
		return num, err
	case AMF0_BOOLEAN_MARKER:
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		return bool(b != 0), nil
	case AMF0_STRING_MARKER:
		return ReadUTF8(r)
	case AMF0_OBJECT_MARKER:
		return ReadObjectProperty(r)
	case AMF0_MOVIECLIP_MARKER:
		return nil, errors.New("Unsupported type: movie clip")
	case AMF0_NULL_MARKER:
		return nil, nil
	case AMF0_UNDEFINED_MARKER:
		return Undefined{}, nil
	case AMF0_REFERENCE_MARKER:
		return nil, errors.New("Unsupported type: reference")
	case AMF0_ECMA_ARRAY_MARKER:
		// Decode ECMA Array to object
		arrLen := make([]byte, 4)
		_, err = r.Read(arrLen)
		if err != nil {
			return nil, err
		}
		obj, err := ReadObjectProperty(r)
		if err != nil {
			return nil, err
		}
		return obj, nil
	case AMF0_OBJECT_END_MARKER:
		return nil, errors.New("Marker error, Object end")
	case AMF0_STRICT_ARRAY_MARKER:
		return nil, errors.New("Unsupported type: strict array")
	case AMF0_DATE_MARKER:
		return nil, errors.New("Unsupported type: date")
	case AMF0_LONG_STRING_MARKER:
		return ReadUTF8Long(r)
	case AMF0_UNSUPPORTED_MARKER:
		return nil, errors.New("Unsupported type: unsupported")
	case AMF0_RECORDSET_MARKER:
		return nil, errors.New("Unsupported type: recordset")
	case AMF0_XML_DOCUMENT_MARKER:
		return nil, errors.New("Unsupported type: XML document")
	case AMF0_TYPED_OBJECT_MARKER:
		return nil, errors.New("Unsupported type: typed object")
	case AMF0_ACMPLUS_OBJECT_MARKER:
		return AMF3_ReadValue(r)
	}
	return nil, errors.New(fmt.Sprintf("Unknown marker type: %d", marker))
}

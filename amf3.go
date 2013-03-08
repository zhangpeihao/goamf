// Copyright 2013, zhangpeihao All rights reserved.

package amf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"sort"
)

//-----------------------------------------------------------------------
// AMF3 Write functions
func AMF3_WriteU29(w Writer, n uint32) (num int, err error) {
	if n <= 0x0000007F {
		err = w.WriteByte(byte(n))
		if err != nil {
			return 0, err
		} else {
			return 1, nil
		}
	} else if n <= 0x00003FFF {
		return w.Write([]byte{byte(n>>7 | 0x80), byte(n & 0x7F)})
	} else if n <= 0x001FFFFF {
		return w.Write([]byte{byte(n>>14 | 0x80), byte(n>>7&0x7F | 0x80), byte(n & 0x7F)})
	} else if n <= 0x1FFFFFFF {
		return w.Write([]byte{byte(n>>22 | 0x80), byte(n>>15&0x7F | 0x80), byte(n>>8&0x7F | 0x80), byte(n)})
	}
	return 0, errors.New("out of range")
}

func AMF3_WriteString(w Writer, str string) (n int, err error) {
	err = w.WriteByte(AMF3_STRING_MARKER)
	if err != nil {
		return 0, err
	}

	n, err = AMF3_WriteUTF8(w, str)
	if err != nil {
		return 1, err
	}
	return 1 + n, nil
}

func AMF3_WriteUTF8(w Writer, str string) (num int, err error) {
	length := len(str)
	if length == 0 {
		err = w.WriteByte(0x01)
		if err != nil {
			return 0, err
		} else {
			return 1, nil
		}
	}
	u := uint32((length << 1) | 0x01) // Todo: reference
	n, err := AMF3_WriteU29(w, u)
	if err != nil {
		return 0, err
	}
	m, err := w.Write([]byte(str))
	if err != nil {
		return n, err
	}
	return m + n, nil
}

func AMF3_WriteDouble(w Writer, num float64) (n int, err error) {
	err = w.WriteByte(AMF3_DOUBLE_MARKER)
	if err != nil {
		return 0, err
	}
	err = binary.Write(w, binary.BigEndian, num)
	if err != nil {
		return 1, err
	}
	return 9, nil
}

func AMF3_WriteBoolean(w Writer, b bool) (n int, err error) {
	if b {
		err = w.WriteByte(AMF3_TRUE_MARKER)
	} else {
		err = w.WriteByte(AMF3_FALSE_MARKER)
	}
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func AMF3_WriteNull(w Writer) (n int, err error) {
	err = w.WriteByte(AMF3_NULL_MARKER)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func AMF3_WriteUndefined(w Writer) (n int, err error) {
	err = w.WriteByte(AMF3_UNDEFINED_MARKER)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func AMF3_WriteObjectMarker(w Writer) (n int, err error) {
	return WriteMarker(w, AMF3_OBJECT_MARKER)
}

func AMF3_WriteObjectEndMarker(w Writer) (n int, err error) {
	err = w.WriteByte(0x01) // Empty string
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func AMF3_WriteObjectName(w Writer, name string) (n int, err error) {
	return AMF3_WriteUTF8(w, name)
}

// Object's item order is uncertainty.
func AMF3_WriteObject(w Writer, obj Object) (n int, err error) {
	n, err = AMF3_WriteObjectMarker(w)
	if err != nil {
		return
	}
	m := 0
	// Write traits flag, Todo: traits class support
	err = w.WriteByte(0x0b)
	if err != nil {
		return
	}
	n += 1
	// Write empty class name
	m, err = AMF3_WriteUTF8(w, "")
	if err != nil {
		return
	}
	n += m
	for key, value := range obj {
		m, err = AMF3_WriteObjectName(w, key)
		if err != nil {
			return
		}
		n += m
		m, err = AMF3_WriteValue(w, value)
		if err != nil {
			return
		}
		n += m
	}
	m, err = AMF3_WriteObjectEndMarker(w)
	return n + m, err
}

func AMF3_WriteValue(w Writer, value interface{}) (n int, err error) {
	if value == nil {
		return AMF3_WriteNull(w)
	}
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return AMF3_WriteNull(w)
	}
	switch v.Kind() {
	case reflect.String:
		return AMF3_WriteString(w, value.(string))
	case reflect.Bool:
		return AMF3_WriteBoolean(w, v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return AMF3_WriteDouble(w, float64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return AMF3_WriteDouble(w, float64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		return AMF3_WriteDouble(w, v.Float())
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// Byte array
			n, err = WriteMarker(w, AMF3_BYTEARRAY_MARKER)
			if err != nil {
				return
			}
			b := v.Bytes()
			length := len(b)
			u := uint32((length << 1) | 0x01) // Todo: reference
			var m int
			m, err = AMF3_WriteU29(w, u)
			if err != nil {
				return
			}
			n += m
			m, err = w.Write(b)
			if err != nil {
				return
			}
			n += m
			return
		} else {
			// Array
			// Dense array only
			n, err = WriteMarker(w, AMF3_ARRAY_MARKER)
			if err != nil {
				return
			}
			// Write dense array length
			length := v.Len()
			u := uint32((length << 1) | 0x01) // Todo: reference
			var m int
			m, err = AMF3_WriteU29(w, u)
			if err != nil {
				return
			}
			n += m
			// Todo: Associative array
			// Empty string to end associative array
			err = w.WriteByte(0x01)
			if err != nil {
				return
			}
			n += 1
			for i := 0; i < length; i++ {
				m, err = AMF3_WriteValue(w, v.Index(i).Interface())
				if err != nil {
					return
				}
				n += m
			}
			return
		}
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return 0, errors.New("Unsupported type")
		}
		n, err = AMF3_WriteObjectMarker(w)
		if err != nil {
			return
		}
		m := 0
		// Write traits flag, Todo: traits class support
		err = w.WriteByte(0x0b)
		if err != nil {
			return
		}
		n += 1
		// Write empty class name
		m, err = AMF3_WriteUTF8(w, "")
		if err != nil {
			return
		}
		n += m

		var sv stringValues = v.MapKeys()
		sort.Sort(sv)
		for _, k := range sv {
			m, err = AMF3_WriteObjectName(w, k.String())
			if err != nil {
				return
			}
			n += m
			m, err = AMF3_WriteValue(w, v.MapIndex(k).Interface())
			if err != nil {
				return
			}
			n += m
		}

		m, err = AMF3_WriteObjectEndMarker(w)
		return n + m, err

	}
	if _, ok := value.(Undefined); ok {
		return AMF3_WriteUndefined(w)
	} else if vt, ok := value.(Object); ok {
		return AMF3_WriteObject(w, vt)
	} else if vt, ok := value.([]interface{}); ok {
		fmt.Printf("Todo: WriteValue: %+v\n", vt)
	}
	return 0, errors.New("Unsupported type")
}

//-----------------------------------------------------------------------
// AMF3 Read functions
func AMF3_ReadU29(r Reader) (n uint32, err error) {
	var b byte
	for i := 0; i < 3; i++ {
		b, err = r.ReadByte()
		if err != nil {
			return
		}
		n = (n << 7) + uint32(b&0x7F)
		if (b & 0x80) == 0 {
			return
		}
	}
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	return ((n << 8) + uint32(b)), nil
}

func AMF3_ReadUTF8(r Reader) (string, error) {
	var length uint32
	var err error
	length, err = AMF3_ReadU29(r)
	if err != nil {
		return "", err
	}
	if length&uint32(0x01) != uint32(1) {
		// Todo: reference
		return "", errors.New("AMF3 reference unsupported")
	}
	length = length >> 1
	if length == 0 {
		return "", nil
	}
	data := make([]byte, length)
	_, err = r.Read(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func AMF3_ReadString(r Reader) (str string, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return "", err
	}
	if marker != AMF3_STRING_MARKER {
		return "", errors.New("Type error")
	}
	return AMF3_ReadUTF8(r)
}

func AMF3_ReadInteger(r Reader) (num uint32, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return 0, err
	}
	if marker != AMF3_INTEGER_MARKER {
		return 0, errors.New("Type error")
	}
	return AMF3_ReadU29(r)
}

func AMF3_ReadDouble(r Reader) (num float64, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return 0, err
	}
	if marker != AMF3_DOUBLE_MARKER {
		return 0, errors.New("Type error")
	}
	err = binary.Read(r, binary.BigEndian, &num)
	return
}

func AMF3_ReadObjectName(r Reader) (name string, err error) {
	return AMF3_ReadUTF8(r)
}

func AMF3_ReadObject(r Reader) (obj Object, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return nil, err
	}
	if marker != AMF3_OBJECT_MARKER {
		return nil, errors.New("Type error")
	}
	return AMF3_ReadObjectProperty(r)
}

func AMF3_ReadObjectProperty(r Reader) (Object, error) {
	obj := make(Object)
	// Read traits flag
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 0x0b {
		return nil, errors.New("Unsupported type: traits object")
	}
	// Read empty string
	b, err = r.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 0x01 {
		return nil, errors.New("Unsupported type: traits object")
	}
	for {
		name, err := AMF3_ReadObjectName(r)
		if err != nil {
			return nil, err
		}
		if name == "" {
			break
		}
		if _, ok := obj[name]; ok {
			return nil, errors.New("object-property exists")
		}
		value, err := AMF3_ReadValue(r)
		if err != nil {
			return nil, err
		}
		obj[name] = value
	}
	return obj, nil
}

func AMF3_ReadByteArray(r Reader) ([]byte, error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return nil, err
	}
	if marker != AMF3_BYTEARRAY_MARKER {
		return nil, errors.New("Type error")
	}
	return AMF3_readByteArray(r)
}

func AMF3_readByteArray(r Reader) ([]byte, error) {
	length, err := AMF3_ReadU29(r)
	if err != nil {
		return nil, err
	}
	if length&uint32(0x01) != uint32(0x01) {
		return nil, errors.New("Unsupported type: reference")
	}
	length = (length >> 1)
	buf := make([]byte, length)
	n, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	if n != int(length) {
		return nil, errors.New("Read buffer size error")
	}
	return buf, nil
}

func AMF3_ReadValue(r Reader) (value interface{}, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return 0, err
	}
	switch marker {
	case AMF3_UNDEFINED_MARKER:
		return Undefined{}, nil
	case AMF3_NULL_MARKER:
		return nil, nil
	case AMF3_FALSE_MARKER:
		return false, nil
	case AMF3_TRUE_MARKER:
		return true, nil
	case AMF3_INTEGER_MARKER:
		return AMF3_ReadU29(r)
	case AMF3_DOUBLE_MARKER:
		var num float64
		err = binary.Read(r, binary.BigEndian, &num)
		return num, err
	case AMF3_STRING_MARKER:
		return AMF3_ReadUTF8(r)
	case AMF3_ARRAY_MARKER:
		// Todo: read array
	case AMF3_OBJECT_MARKER:
		return AMF3_ReadObjectProperty(r)
	case AMF3_BYTEARRAY_MARKER:
		return AMF3_readByteArray(r)
	}

	return nil, errors.New(fmt.Sprintf("Unknown marker type: %d", marker))
}

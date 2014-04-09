package amf

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestWriteMarker(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := WriteMarker(buf, AMF0_NUMBER_MARKER)
	if err != nil {
		t.Errorf("test %s err", "WriteMark", err)
	} else {
		expect := []byte{0x00}
		got := buf.Bytes()
		if !bytes.Equal(expect, got) {
			t.Errorf("bytes: expect %x got %x", expect, got)
		}
		if n != 1 {
			t.Errorf("n: expect %x got %x", expect, got)
		}
	}
}

func TestWriteUTF8(t *testing.T) {
	buf := new(bytes.Buffer)
	err := WriteUTF8(buf, "foo", uint16(3))
	if err != nil {
		t.Errorf("test for %s error: %s", "foo", err)
	} else {
		expect := []byte{0x00, 0x03, 'f', 'o', 'o'}
		got := buf.Bytes()
		if !bytes.Equal(expect, got) {
			t.Errorf("expect %x got %x", expect, got)
		}
	}
	buf = new(bytes.Buffer)
	err = WriteUTF8(buf, "你好", uint16(len("你好")))
	if err != nil {
		t.Errorf("test for %s error: %s", "你好", err)
	} else {
		expect := []byte{0x00, 0x06, 0xe4, 0xbd, 0xa0, 0xe5, 0xa5, 0xbd}
		got := buf.Bytes()
		if !bytes.Equal(expect, got) {
			t.Errorf("expect %x got %x", expect, got)
		}
	}
}

func TestWriteUTF8Long(t *testing.T) {
	buf := new(bytes.Buffer)
	testBytes := []byte("12345678")
	longStringBuf := new(bytes.Buffer)
	for i := 0; i < 65536; i++ {
		longStringBuf.Write(testBytes)
	}
	err := WriteUTF8Long(buf, string(longStringBuf.Bytes()), uint32(longStringBuf.Len()))
	if err != nil {
		t.Errorf("test for long string error: %s", err)
	} else {
		var length uint32
		err = binary.Read(buf, binary.BigEndian, &length)
		if err != nil {
			t.Fatal("test long string result check, read length error:", err)
		}
		if length != (65536 * 8) {
			t.Errorf("String length error: %d\n", length)
		}
		tmpBuf := make([]byte, 8)
		counter := 0
		for buf.Len() > 0 {
			n, err := buf.Read(tmpBuf)
			if err != nil {
				t.Fatalf("test long string result check, read data(%d) error: %s, n: %d", counter, err.Error(), n)
			}
			if n != 8 {
				t.Fatalf("test long string result check, read data(%d) n: %d", counter, n)
			}
			if !bytes.Equal(testBytes, tmpBuf) {
				t.Fatalf("test long string result check, read data % x", tmpBuf)
			}

			counter++
		}
		if counter != 65536 {
			t.Errorf("test for long string result check, counter is %d", counter)
		}
	}
}

func TestEncodeString(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := WriteString(buf, "foo")
	if err != nil {
		t.Fatalf("WriteString error: %s", err)
	}
	if n != 6 {
		t.Errorf("WriteString return n: %d\n", n)
	}
	expect := []byte{0x02, 0x00, 0x03, 0x66, 0x6f, 0x6f}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}
func TestEncodeLongString(t *testing.T) {
	testBytes := []byte("12345678")
	longStringBuf := new(bytes.Buffer)
	for i := 0; i < 65536; i++ {
		longStringBuf.Write(testBytes)
	}
	buf := new(bytes.Buffer)
	n, err := WriteString(buf, string(longStringBuf.Bytes()))
	if err != nil {
		t.Fatalf("WriteLongString error: %s", err)
	}
	if n != 1+4+(65536*8) {
		t.Errorf("WriteString return n: %d\n", n)
	}
	marker, err := buf.ReadByte()
	if err != nil {
		t.Fatal("Read marker error:", err)
	}
	if marker != AMF0_LONG_STRING_MARKER {
		t.Fatalf("Read marker is:0x%02x", marker)
	}
	var length uint32
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		t.Fatal("test long string result check, read length error:", err)
	}
	if length != (65536 * 8) {
		t.Errorf("String length error: %d\n", length)
	}
	tmpBuf := make([]byte, 8)
	counter := 0
	for buf.Len() > 0 {
		n, err := buf.Read(tmpBuf)
		if err != nil {
			t.Fatalf("test long string result check, read data(%d) error: %s, n: %d", counter, err.Error(), n)
		}
		if n != 8 {
			t.Fatalf("test long string result check, read data(%d) n: %d", counter, n)
		}
		if !bytes.Equal(testBytes, tmpBuf) {
			t.Fatalf("test long string result check, read data % x", tmpBuf)
		}

		counter++
	}
	if counter != 65536 {
		t.Errorf("test for long string result check, counter is %d", counter)
	}
}

func TestEncodeFloat(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := WriteDouble(buf, float64(1.2))
	if err != nil {
		t.Fatalf("WriteDouble error: %s", err)
	}
	if n != 9 {
		t.Errorf("WriteDouble return n: %d\n", n)
	}
	if buf.Len() != 9 {
		t.Errorf("WriteDouble writen buffer len: %d\n", buf.Len())
	}
	expect := []byte{0x00, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Fatalf("expect % x got % x", expect, got)
	}
}

func TestEncodeBoolean(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := WriteBoolean(buf, true)
	if err != nil {
		t.Fatalf("WriteBoolean error: %s", err)
	}
	if n != 2 {
		t.Errorf("WriteBoolean return n: %d\n", n)
	}
	expect := []byte{0x01, 0x01}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

func TestEncodeNull(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := WriteNull(buf)
	if err != nil {
		t.Fatalf("WriteNull error: %s", err)
	}
	if n != 1 {
		t.Errorf("WriteNull return n: %d\n", n)
	}
	expect := []byte{0x05}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

func TestEncodeUndefined(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := WriteUndefined(buf)
	if err != nil {
		t.Fatalf("WriteUndefined error: %s", err)
	}
	if n != 1 {
		t.Errorf("WriteUndefined return n: %d\n", n)
	}
	expect := []byte{0x06}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("expect %x got %x", expect, got)
	}
}

type SubStruct struct {
	data string `amf:"data"`
}

type Embedded struct {
	member string `amf:"member"`
}

type Struct struct {
	Embedded
	Name   string     `amf:"name"`
	Sub    *SubStruct `amf:"sub"`
	Empty  string
	Unused string `amf:"-"`
}

type TestEncodeValueCase struct {
	name   string
	v      interface{}
	expect []byte
}

var testCases = []TestEncodeValueCase{
	{"1.2", 1.2, []byte{0x00, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}},
	{"float32(1.2)", float32(1.2), []byte{0x00, 0x3f, 0xf3, 0x33, 0x33, 0x40, 0x00, 0x00, 0x00}},
	{"float64(1.2)", float64(1.2), []byte{0x00, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}},
	{"1", 1, []byte{0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"1", int(1), []byte{0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"1", int8(1), []byte{0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"1", int16(1), []byte{0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"1", int32(1), []byte{0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"1", int64(1), []byte{0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"1", uint(1), []byte{0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"1", uint8(1), []byte{0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"1", uint16(1), []byte{0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"1", uint32(1), []byte{0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"1", uint64(1), []byte{0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"-1", int(-1), []byte{0x00, 0xbf, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"-1", int8(-1), []byte{0x00, 0xbf, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"-1", int16(-1), []byte{0x00, 0xbf, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"-1", int32(-1), []byte{0x00, 0xbf, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"-1", int64(-1), []byte{0x00, 0xbf, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"foo", "foo", []byte{0x02, 0x00, 0x03, 'f', 'o', 'o'}},
	{"empty string", "", []byte{0x02, 0x00, 0x00}},
	{"false", false, []byte{0x01, 0x00}},
	{"true", true, []byte{0x01, 0x01}},
	{"null", nil, []byte{0x05}},
	{"array", []string{"a", "b", "c"},
		[]byte{0x08,
			0x00, 0x00, 0x00, 0x03,
			0x00, 0x01, '0', 0x02, 0x00, 0x01, 'a',
			0x00, 0x01, '1', 0x02, 0x00, 0x01, 'b',
			0x00, 0x01, '2', 0x02, 0x00, 0x01, 'c',
			0x00, 0x00, 0x09,
		}},
	{"struct", &Struct{Embedded{"emb"}, "zhang", &SubStruct{"123"}, "noname", "unused"},
		[]byte{0x03,
			0x00, 0x06, 'm', 'e', 'm', 'b', 'e', 'r', 0x02, 0x00, 0x03, 'e', 'm', 'b', // member: emb [Embedded]
			0x00, 0x04, 'n', 'a', 'm', 'e', 0x02, 0x00, 0x05, 'z', 'h', 'a', 'n', 'g', // name: zhang
			0x00, 0x03, 's', 'u', 'b', 0x03, // sub: SubStruct{
			0x00, 0x04, 'd', 'a', 't', 'a', 0x02, 0x00, 0x03, '1', '2', '3', // data: 123
			0x00, 0x00, 0x09, // }
			0x00, 0x05, 'E', 'm', 'p', 't', 'y', 0x02, 0x00, 0x06, 'n', 'o', 'n', 'a', 'm', 'e', // Empty: noname
			0x00, 0x00, 0x09,
		}},
	/*
		{"object", Object{"0": "a", "1": "b", "2": "c"},
			[]byte{0x03,
				0x00, 0x01, '1', 0x02, 0x00, 0x01, 'b',
				0x00, 0x01, '0', 0x02, 0x00, 0x01, 'a',
				0x00, 0x01, '2', 0x02, 0x00, 0x01, 'c',
				0x00, 0x00, 0x09,
			}},
	*/
	{"Map", map[string]string{"0": "a", "1": "b", "2": "c"},
		[]byte{0x03,
			0x00, 0x01, '0', 0x02, 0x00, 0x01, 'a',
			0x00, 0x01, '1', 0x02, 0x00, 0x01, 'b',
			0x00, 0x01, '2', 0x02, 0x00, 0x01, 'c',
			0x00, 0x00, 0x09,
		}},
}

func TestEncodeValue(t *testing.T) {
	for _, c := range testCases {
		buf := new(bytes.Buffer)
		n, err := WriteValue(buf, c.v)
		if err != nil {
			t.Errorf("WriteValue(%s) error: %s", c.name, err.Error())
			continue
		}
		if n != len(c.expect) {
			t.Errorf("WriteValue(%s) return n: %d, expect %d\n", c.name, n, len(c.expect))

		}
		got := buf.Bytes()
		if !bytes.Equal(c.expect, got) {
			t.Errorf("WriteValue(%s)\n   got: % 2x\nexpect: % 2x\n", c.name, got, c.expect)
			continue
		}
	}
}

func TestEncodeObject(t *testing.T) {
	buf := new(bytes.Buffer)
	obj := Object{"0": "a", "1": "b", "2": "c"}
	n, err := WriteObject(buf, obj)
	if err != nil {
		t.Fatalf("Object %s", err)
	}
	if n != 25 {
		t.Errorf("WriteObject return n: %d\n", n)
	}
	if buf.Len() != 25 {
		t.Errorf("WriteDouble writen buffer len: %d\n", buf.Len())
	}
	/*
		expect := []byte{0x03,
			0x00, 0x01, '1', 0x02, 0x00, 0x01, 'b',
			0x00, 0x01, '0', 0x02, 0x00, 0x01, 'a',
			0x00, 0x01, '2', 0x02, 0x00, 0x01, 'c',
			0x00, 0x00, 0x09,
		}
	*/
	got := buf.Bytes()
	if got[0] != 0x03 {
		t.Errorf("got[0]: %x, expect %x\n", got[0], 0x03)
	}
	names := make([]string, len(obj))
	i := 0
	for k, _ := range obj {
		names[i] = k
		i++
	}
	for i = 0; i < 3; i++ {
		if got[1+i*7] != 0x00 || got[1+i*7+1] != 0x01 {
			t.Errorf("item[%d] name len: %x, expect 0001\n", i, got[(1+i*7):(1+i*7+2)])
		}
		name := string(got[1+i*7+2 : 1+i*7+3])
		found := false
		for index, n := range names {
			if n == name {
				names = append(names[:index], names[index+1:]...)
				found = true
				break
			}
		}
		if !found {
			t.Errorf("item[%d] name: %s not in expected slices: %v\n", i, name, names)
		}
		if got[1+i*7+3] != 0x02 {
			t.Errorf("item[%d] marker: %x, expect 02\n", i, got[(1+i*7+3):(1+i*7+4)])
		}
		if got[1+i*7+4] != 0x00 || got[1+i*7+5] != 0x01 {
			t.Errorf("item[%d] value len: %x, expect 0001\n", i, got[(1+i*7+4):(1+i*7+6)])
		}
		value := string(got[1+i*7+6 : 1+i*7+7])
		if obj[name] != value {
			t.Errorf("item[%d] value: %s, expect: %s\n", i, value, obj[name])

		}
	}
}

//-------------------------------------------------------

func TestReadMarker(t *testing.T) {
	buf := bytes.NewBuffer([]byte{AMF0_NUMBER_MARKER})
	marker, err := ReadMarker(buf)
	if err != nil {
		t.Error("ReadMarker err:", err)
	} else {
		if AMF0_NUMBER_MARKER != marker {
			t.Errorf("ReadMarker: expect %x got %x", AMF0_NUMBER_MARKER, marker)
		}
	}
}

func TestReadString(t *testing.T) {
	buf := bytes.NewReader([]byte{0x02, 0x00, 0x03, 0x66, 0x6f, 0x6f})
	expect := "foo"
	got, err := ReadString(buf)
	if err != nil {
		t.Fatalf("ReadString error: %s", err)
	}
	if expect != got {
		t.Fatalf("expect %v got %v", expect, got)
	}
}

func TestReadLongString(t *testing.T) {
	buf := bytes.NewReader([]byte{0x0c, 0x00, 0x00, 0x00, 0x03, 0x66, 0x6f, 0x6f})
	expect := "foo"
	got, err := ReadString(buf)
	if err != nil {
		t.Fatalf("ReadLongString error: %s", err)
	}
	if expect != got {
		t.Fatalf("expect %v got %v", expect, got)
	}
}

func TestDecodeNumber(t *testing.T) {
	buf := bytes.NewReader([]byte{0x00, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33})
	got, err := ReadDouble(buf)
	if err != nil {
		t.Fatalf("ReadDouble error: %s", err)
	}
	expect := float64(1.2)
	if expect != got {
		t.Fatalf("expect %v got %v", expect, got)
	}
}

func TestDecodeBoolean(t *testing.T) {
	buf := bytes.NewReader([]byte{0x01, 0x01})
	expect := true
	got, err := ReadBoolean(buf)
	if err != nil {
		t.Fatalf("ReadBoolean error: %s", err)
	}
	if expect != got {
		t.Fatalf("expect %v got %v", expect, got)
	}
}

var testDecodeCases = []TestEncodeValueCase{
	{"1.2", 1.2, []byte{0x00, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}},
	{"float64(1.2)", float64(1.2), []byte{0x00, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}},
	{"1", 1.0, []byte{0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	{"foo", "foo", []byte{0x02, 0x00, 0x03, 'f', 'o', 'o'}},
	{"empty string", "", []byte{0x02, 0x00, 0x00}},
	{"false", false, []byte{0x01, 0x00}},
	{"true", true, []byte{0x01, 0x01}},
	{"null", nil, []byte{0x05}},
}

func TestDecodeValue(t *testing.T) {
	for _, c := range testDecodeCases {
		buf := bytes.NewReader(c.expect)
		value, err := ReadValue(buf)
		if err != nil {
			t.Errorf("ReadValue(%s) error: %s", c.name, err.Error())
			continue
		}
		if value != c.v {
			t.Errorf("ReadValue(%s) return %v, expect %v\n", c.name, value, c.v)

		}
	}
}

func TestDecodeObject(t *testing.T) {
	buf := bytes.NewReader([]byte{0x03,
		0x00, 0x01, '1', 0x02, 0x00, 0x01, 'b',
		0x00, 0x01, '0', 0x02, 0x00, 0x01, 'a',
		0x00, 0x01, '2', 0x02, 0x00, 0x01, 'c',
		0x00, 0x00, 0x09,
	})
	obj, err := ReadObject(buf)
	if err != nil {
		t.Fatalf("ReadObject error: %s", err)
	}
	if len(obj) != 3 {
		t.Errorf("ReadObject return len :%d, expect 3\n", len(obj))
	}
	expect := Object{
		"0": "a",
		"1": "b",
		"2": "c",
	}
	for k, v := range obj {
		if expectV, ok := expect[k]; ok {
			if v != expectV {
				t.Errorf("ReadObject item[%s] value unmatck! got %s, exprct %s", k, v, expectV)
			}
			delete(expect, k)
		} else {
			t.Errorf("ReadObject not found key: %s", k)
		}
	}
	if len(expect) != 0 {
		t.Errorf("ReadObject loss some items: %v", expect)
	}
}

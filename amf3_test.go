package amf

import (
	"bytes"
	"testing"
)

type testU29Case struct {
	value  uint32
	expect []byte
}

var testU29Cases = []testU29Case{
	{1, []byte{0x01}},
	{2, []byte{0x02}},
	{127, []byte{0x7F}},
	{128, []byte{0x81, 0x00}},
	{255, []byte{0x81, 0x7F}},
	{256, []byte{0x82, 0x00}},
	{0x3FFF, []byte{0xFF, 0x7F}},
	{0x4000, []byte{0x81, 0x80, 0x00}},
	{0x7FFF, []byte{0x81, 0xFF, 0x7F}},
	{0x8000, []byte{0x82, 0x80, 0x00}},
	{0x1FFFFF, []byte{0xFF, 0xFF, 0x7F}},
	{0x200000, []byte{0x80, 0xC0, 0x80, 0x00}},
	{0x3FFFFF, []byte{0x80, 0xFF, 0xFF, 0xFF}},
	{0x400000, []byte{0x81, 0x80, 0x80, 0x00}},
	{0x0FFFFFFF, []byte{0xBF, 0xFF, 0xFF, 0xFF}},
}

func TestWriteU29(t *testing.T) {
	for _, c := range testU29Cases {
		buf := new(bytes.Buffer)
		n, err := AMF3_WriteU29(buf, c.value)
		if err != nil {
			t.Errorf("AMF3_WriteU29(%d) error: %s", c.value, err)
		} else {
			if n != len(c.expect) {
				t.Errorf("AMF3_WriteU29 expect n %x got %x", len(c.expect), n)
			}
			got := buf.Bytes()
			if !bytes.Equal(c.expect, got) {
				t.Errorf("AMF3_WriteU29 expect buffer %x got %x", c.expect, got)
			}
		}
	}
}

func TestAMF3_WriteUTF8(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := AMF3_WriteUTF8(buf, "foo")
	if err != nil {
		t.Errorf("test for %s error: %s", "foo", err)
	} else {
		if n != 4 {
			t.Errorf("AMF3_WriteU29 expect n %x got %x", 4, n)
		}
		expect := []byte{0x07, 'f', 'o', 'o'}
		got := buf.Bytes()
		if !bytes.Equal(expect, got) {
			t.Errorf("AMF3_WriteUTF8 expect buffer %x got %x", expect, got)
		}
	}
	buf = new(bytes.Buffer)
	n, err = AMF3_WriteUTF8(buf, "你好")
	if err != nil {
		t.Errorf("test for %s error: %s", "你好", err)
	} else {
		if n != 7 {
			t.Errorf("AMF3_WriteU29 expect n %x got %x", 7, n)
		}
		expect := []byte{0x0D, 0xe4, 0xbd, 0xa0, 0xe5, 0xa5, 0xbd}
		got := buf.Bytes()
		if !bytes.Equal(expect, got) {
			t.Errorf("AMF3_WriteUTF8 expect buffer %x got %x", expect, got)
		}
	}
}

func TestAMF3_EncodeString(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := AMF3_WriteString(buf, "foo")
	if err != nil {
		t.Fatalf("AMF3_WriteString error: %s", err)
	}
	if n != 5 {
		t.Errorf("AMF3_WriteString return n: %d\n", n)
	}
	expect := []byte{0x06, 0x07, 'f', 'o', 'o'}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("AMF3_WriteString expect %x got %x", expect, got)
	}
}

func TestAMF3_EncodeFloat(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := AMF3_WriteDouble(buf, float64(1.2))
	if err != nil {
		t.Fatalf("AMF3_WriteDouble error: %s", err)
	}
	if n != 9 {
		t.Errorf("AMF3_WriteDouble return n: %d\n", n)
	}
	if buf.Len() != 9 {
		t.Errorf("AMF3_WriteDouble writen buffer len: %d\n", buf.Len())
	}
	expect := []byte{0x05, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Fatalf("AMF3_WriteDouble expect % x got % x", expect, got)
	}
}

func TestAMF3_EncodeBoolean(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := AMF3_WriteBoolean(buf, true)
	if err != nil {
		t.Fatalf("AMF3_WriteBoolean error: %s", err)
	}
	if n != 1 {
		t.Errorf("AMF3_WriteBoolean return n: %d\n", n)
	}
	expect := []byte{0x03}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("AMF3_WriteBoolean(true) expect %x got %x", expect, got)
	}

	buf = new(bytes.Buffer)
	n, err = AMF3_WriteBoolean(buf, false)
	if err != nil {
		t.Fatalf("AMF3_WriteBoolean error: %s", err)
	}
	if n != 1 {
		t.Errorf("AMF3_WriteBoolean return n: %d\n", n)
	}
	expect = []byte{0x02}
	got = buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("AMF3_WriteBoolean(false) expect %x got %x", expect, got)
	}
}

func TestAMF3_EncodeNull(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := AMF3_WriteNull(buf)
	if err != nil {
		t.Fatalf("AMF3_WriteNull error: %s", err)
	}
	if n != 1 {
		t.Errorf("AMF3_WriteNull return n: %d\n", n)
	}
	expect := []byte{0x01}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("AMF3_WriteNull expect %x got %x", expect, got)
	}
}

func TestAMF3_EncodeUndefined(t *testing.T) {
	buf := new(bytes.Buffer)
	n, err := AMF3_WriteUndefined(buf)
	if err != nil {
		t.Fatalf("AMF3_WriteUndefined error: %s", err)
	}
	if n != 1 {
		t.Errorf("AMF3_WriteUndefined return n: %d\n", n)
	}
	expect := []byte{0x00}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("AMF3_WriteUndefined expect %x got %x", expect, got)
	}
}

func TestAMF3_EncodeObject(t *testing.T) {
	buf := new(bytes.Buffer)
	obj := Object{"0": "a", "1": "b", "2": "c"}
	n, err := AMF3_WriteObject(buf, obj)
	if err != nil {
		t.Fatalf("AMF3_WriteObject %s", err)
	}
	if n != 19 {
		t.Errorf("AMF3_WriteObject return n: %d\n", n)
	}
	if buf.Len() != 19 {
		t.Errorf("AMF3_WriteDouble writen buffer len: %d\n", buf.Len())
	}
	/*
		expect := []byte{0x0A, 0x0B, 0x01
			0x03, '0', 0x06, 0x03, 'a',
			0x03, '1', 0x06, 0x03, 'b',
			0x03, '2', 0x06, 0x03, 'c',
			0x01,
		}
	*/
	got := buf.Bytes()
	if got[0] != 0x0A || got[1] != 0x0B || got[2] != 0x01 {
		t.Errorf("AMF3_WriteObject got[:3]: %x, expect %x\n", got[:3], []byte{0x0A, 0x0B, 0x01})
	}
	names := make([]string, len(obj))
	i := 0
	for k, _ := range obj {
		names[i] = k
		i++
	}
	for i = 0; i < 3; i++ {
		if got[3+i*5] != 0x03 {
			t.Errorf("AMF3_WriteObject item[%d] name len: 0x%x, expect 0x03\n", i, got[3+i*5])
		}
		name := string(got[3+i*5+1 : 3+i*5+2])
		found := false
		for index, n := range names {
			if n == name {
				names = append(names[:index], names[index+1:]...)
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AMF3_WriteObject item[%d] name: %s not in expected slices: %v\n", i, name, names)
		}
		if got[3+i*5+2] != 0x06 || got[3+i*5+3] != 0x03 {
			t.Errorf("AMF3_WriteObject item[%d] marker: %x, expect 0603\n", i, got[(3+i*5+2):(3+i*5+4)])
		}
		value := string(got[(3 + i*5 + 4):(3 + i*5 + 5)])
		if obj[name] != value {
			t.Errorf("AMF3_WriteObject item[%d] value: %s, expect: %s\n", i, value, obj[name])

		}
	}
}

func TestAMF3_EncodeByteArray(t *testing.T) {
	buf := new(bytes.Buffer)
	b := []byte("foo")
	n, err := AMF3_WriteValue(buf, b)
	if err != nil {
		t.Fatalf("TestAMF3_EncodeByteArray error: %s", err)
	}
	if n != 5 {
		t.Errorf("TestAMF3_EncodeByteArray return n: %d\n", n)
	}
	expect := []byte{0x0c, 0x07, 'f', 'o', 'o'}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("TestAMF3_EncodeByteArray expect %x got %x", expect, got)
	}
}

func TestAMF3_EncodeArray(t *testing.T) {
	buf := new(bytes.Buffer)
	a := []string{
		"1", "2", "3", "4", "5", "6", "7", "8", "9",
	}
	n, err := AMF3_WriteValue(buf, a)
	if err != nil {
		t.Fatalf("TestAMF3_EncodeArray error: %s", err)
	}
	if n != 30 {
		t.Errorf("TestAMF3_EncodeArray return n: %d\n", n)
	}
	expect := []byte{0x09, 0x13, 0x01,
		0x06, 0x03, '1',
		0x06, 0x03, '2',
		0x06, 0x03, '3',
		0x06, 0x03, '4',
		0x06, 0x03, '5',
		0x06, 0x03, '6',
		0x06, 0x03, '7',
		0x06, 0x03, '8',
		0x06, 0x03, '9',
	}
	got := buf.Bytes()
	if !bytes.Equal(expect, got) {
		t.Errorf("TestAMF3_EncodeArray expect %x got %x", expect, got)
	}
}

//-------------------------------------------------------
func TestAMF3_ReadU29(t *testing.T) {
	for _, c := range testU29Cases {
		buf := bytes.NewBuffer(c.expect)
		n, err := AMF3_ReadU29(buf)
		if err != nil {
			t.Errorf("AMF3_ReadU29(%d) error: %s", c.value, err)
		} else {
			if n != c.value {
				t.Errorf("AMF3_WriteU29 expect n %x got %x", c.value, n)
			}
		}
	}
}

func TestAMF3_ReadUTF8(t *testing.T) {
	expect := "foo"
	buf := bytes.NewBuffer([]byte{0x07, 'f', 'o', 'o'})
	got, err := AMF3_ReadUTF8(buf)
	if err != nil {
		t.Errorf("TestAMF3_ReadUTF8 test for %s error: %s", expect, err)
	} else {
		if got != expect {
			t.Errorf("TestAMF3_ReadUTF8 expect %s got %s", expect, got)
		}
	}
}

func TestAMF3_ReadString(t *testing.T) {
	expect := "foo"
	buf := bytes.NewBuffer([]byte{0x06, 0x07, 'f', 'o', 'o'})
	got, err := AMF3_ReadString(buf)
	if err != nil {
		t.Errorf("TestAMF3_ReadString test for %s error: %s", expect, err)
	} else {
		if got != expect {
			t.Errorf("TestAMF3_ReadString expect %s got %s", expect, got)
		}
	}
}

func TestAMF3_DecodeDouble(t *testing.T) {
	expect := float64(1.2)
	buf := bytes.NewBuffer([]byte{0x05, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33})
	got, err := AMF3_ReadDouble(buf)
	if err != nil {
		t.Fatalf("TestAMF3_DecodeDouble error: %s", err)
	}
	if got != expect {
		t.Errorf("TestAMF3_DecodeDouble got %v, expect %v\n", got, expect)
	}
}

var testAMF3_DecodeCases = []TestEncodeValueCase{
	{"1.2", 1.2, []byte{0x05, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}},
	{"float64(1.2)", float64(1.2), []byte{0x05, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}},
	{"1", uint32(1), []byte{0x04, 0x01}},
	{"foo", "foo", []byte{0x06, 0x07, 'f', 'o', 'o'}},
	{"empty string", "", []byte{0x06, 0x01}},
	{"false", false, []byte{0x02}},
	{"true", true, []byte{0x03}},
	{"null", nil, []byte{0x01}},
}

func TestAMF3_DecodeValue(t *testing.T) {
	for _, c := range testAMF3_DecodeCases {
		buf := bytes.NewReader(c.expect)
		value, err := AMF3_ReadValue(buf)
		if err != nil {
			t.Errorf("AMF3_ReadValue(%s) error: %s", c.name, err.Error())
			continue
		}
		if value != c.v {
			t.Errorf("AMF3_ReadValue(%s) return %v, expect %v\n", c.name, value, c.v)

		}
	}
}

func TestAMF3_DecodeObject(t *testing.T) {
	buf := bytes.NewReader(
		[]byte{0x0A, 0x0B, 0x01,
			0x03, '0', 0x06, 0x03, 'a',
			0x03, '1', 0x06, 0x03, 'b',
			0x03, '2', 0x06, 0x03, 'c',
			0x01,
		})
	obj, err := AMF3_ReadObject(buf)
	if err != nil {
		t.Fatalf("TestAMF3_DecodeObject error: %s", err)
	}
	if len(obj) != 3 {
		t.Errorf("TestAMF3_DecodeObject return len :%d, expect 3\n", len(obj))
	}
	expect := Object{
		"0": "a",
		"1": "b",
		"2": "c",
	}
	for k, v := range obj {
		if expectV, ok := expect[k]; ok {
			if v != expectV {
				t.Errorf("TestAMF3_DecodeObject item[%s] value unmatck! got %s, exprct %s", k, v, expectV)
			}
			delete(expect, k)
		} else {
			t.Errorf("TestAMF3_DecodeObject not found key: %s", k)
		}
	}
	if len(expect) != 0 {
		t.Errorf("TestAMF3_DecodeObject loss some items: %v", expect)
	}
}

func TestAMF3_DecodeByteArray(t *testing.T) {
	buf := bytes.NewReader(
		[]byte{0x0c, 0x07, 'f', 'o', 'o'})
	got, err := AMF3_ReadByteArray(buf)
	if err != nil {
		t.Fatalf("AMF3_ReadByteArray error: %s", err)
	}
	expect := []byte("foo")
	if !bytes.Equal(expect, got) {
		t.Errorf("AMF3_ReadByteArray expect %x got %x", expect, got)
	}

}

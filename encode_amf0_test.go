// Copyright 2013, zhangpeihao All rights reserved.

package amf

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestString(t *testing.T) {
	buf := new(bytes.Buffer)
	encoder := NewAMF0Encoder(buf)
	err := encoder.Encode("123456789")
	if err != nil {
		t.Fatal("error:", err)
	}
	if 0 != bytes.Compare(buf.Bytes(), []byte{0x02, 0x00, 0x09,
		0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39}) {
		t.Errorf("Encode string 123456789 to AMF0 error: result: % 0x\r\n", buf.Bytes())
	}
}

func TestBooleanTrue(t *testing.T) {
	buf := new(bytes.Buffer)
	encoder := NewAMF0Encoder(buf)
	err := encoder.Encode(true)
	if err != nil {
		t.Fatal("error:", err)
	}
	if 0 != bytes.Compare(buf.Bytes(), []byte{0x01, 0x01}) {
		t.Errorf("Encode boolean true to AMF0 error: result: % 0x\r\n", buf.Bytes())
	}
}

func TestBooleanFalse(t *testing.T) {
	buf := new(bytes.Buffer)
	encoder := NewAMF0Encoder(buf)
	err := encoder.Encode(false)
	if err != nil {
		t.Fatal("error:", err)
	}
	if 0 != bytes.Compare(buf.Bytes(), []byte{0x01, 0x00}) {
		t.Errorf("Encode boolean false to AMF0 error: result: % 0x\r\n", buf.Bytes())
	}
}

func TestIntegter(t *testing.T) {
	buf := new(bytes.Buffer)
	encoder := NewAMF0Encoder(buf)
	err := encoder.Encode(1000)
	if err != nil {
		t.Fatal("error:", err)
	}
	if 0 != bytes.Compare(buf.Bytes(), []byte{0x00, 0x40, 0x8f, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00}) {
		t.Errorf("Encode integer 1000 to AMF0 error: result: % 0x\r\n", buf.Bytes())
	}
}

func TestIntegterPtr(t *testing.T) {
	buf := new(bytes.Buffer)
	encoder := NewAMF0Encoder(buf)
	i := 1000
	err := encoder.Encode(&i)
	if err != nil {
		t.Fatal("error:", err)
	}
	if 0 != bytes.Compare(buf.Bytes(), []byte{0x00, 0x40, 0x8f, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00}) {
		t.Errorf("Encode integer point 1000 to AMF0 error: result: % 0x\r\n", buf.Bytes())
	}
}

func TestUnsignedIntegter(t *testing.T) {
	buf := new(bytes.Buffer)
	encoder := NewAMF0Encoder(buf)
	err := encoder.Encode(uint(1000))
	if err != nil {
		t.Fatal("error:", err)
	}
	if 0 != bytes.Compare(buf.Bytes(), []byte{0x00, 0x40, 0x8f, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00}) {
		t.Errorf("Encode unsigned integer 1000 to AMF0 error: result: % 0x\r\n", buf.Bytes())
	}
}

func TestUnsignedIntegterPtr(t *testing.T) {
	buf := new(bytes.Buffer)
	encoder := NewAMF0Encoder(buf)
	i := uint(1000)
	err := encoder.Encode(&i)
	if err != nil {
		t.Fatal("error:", err)
	}
	if 0 != bytes.Compare(buf.Bytes(), []byte{0x00, 0x40, 0x8f, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00}) {
		t.Errorf("Encode unsigned integer point 1000 to AMF0 error: result: % 0x\r\n", buf.Bytes())
	}
}

func TestDouble(t *testing.T) {
	buf := new(bytes.Buffer)
	encoder := NewAMF0Encoder(buf)
	err := encoder.Encode(float64(1234567890.123456789))
	if err != nil {
		t.Fatal("error:", err)
	}
	if 0 != bytes.Compare(buf.Bytes(), []byte{0x00, 0x41, 0xd2, 0x65, 0x80, 0xb4, 0x87, 0xe6, 0xb7}) {
		t.Errorf("Encode integer 1000 to AMF0 error: result: % 0x\r\n", buf.Bytes())
	}
}

type SubStruct struct {
	Sub string `amf:"string"`
}

type Struct struct {
	B   bool           `amf:"bool"`
	I   int            `amf:"int"`
	F   float64        `amf:"float"`
	S   string         `amf:"string"`
	A   []string       `amf:"array"`
	M   map[string]int `amf:"map"`
	BA  []byte         `amf:"bytearray"`
	Sub SubStruct      `amf:"sub"`
}

func TestStruct(t *testing.T) {
	buf := new(bytes.Buffer)
	encoder := NewAMF0Encoder(buf)
	s := Struct{
		B:   true,
		I:   1000,
		F:   0.001,
		S:   "hello world",
		A:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		M:   make(map[string]int),
		BA:  []byte("byte array"),
		Sub: SubStruct{"sub"},
	}
	for i := 0; i < 10; i++ {
		s.M[strconv.Itoa(i)] = i
	}
	err := encoder.Encode(s)
	if err != nil {
		t.Fatal("error:", err)
	}
	if 0 != bytes.Compare(buf.Bytes(), []byte{0x00, 0x41, 0xd2, 0x65, 0x80, 0xb4, 0x87, 0xe6, 0xb7}) {
		t.Errorf("Encode struct to AMF0 error: result: % 0x\r\n", buf.Bytes())
		file, err := os.OpenFile("amf0-obj.bin", os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Println("Dump data failed! Error:", err)
		} else {
			file.Write(buf.Bytes())
			file.Close()
		}
	}
}

// Copyright 2013, zhangpeihao All rights reserved.

package amf

import (
	"io"
)

// An AMF0Encoder writes AMF0 objects to an output stream.
type AMF0Encoder struct {
	w   io.Writer
	e   encodeStateAMF0
	err error
}

// NewAMF0Encoder returns a new encoder that writes to w.
func NewAMF0Encoder(w io.Writer) *AMF0Encoder {
	return &AMF0Encoder{w: w}
}

// Encode writes the AMF0 encoding of v to the connection.
//
// See the documentation for Marshal for details about the
// conversion of Go values to AMF0.
func (enc *AMF0Encoder) Encode(v interface{}) error {
	if enc.err != nil {
		return enc.err
	}
	enc.e.Reset()
	err := enc.e.marshal(v)
	if err != nil {
		return err
	}

	if _, err = enc.w.Write(enc.e.Bytes()); err != nil {
		enc.err = err
	}
	return err
}

//go:build gccgo

package log

import "encoding/base64"

// Base64 adds base64 encoding of the value to the entry.
//
// gccgo's libgo is based on go1.18, which lacks base64.Encoding.AppendEncode,
// so the encoding is written into a scratch buffer first.
func (e *Entry) Base64(key string, value []byte) *Entry {
	if e == nil {
		return nil
	}

	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(value)))
	base64.StdEncoding.Encode(buf, value)
	e.buf = append(e.buf, buf...)
	e.buf = append(e.buf, '"')
	return e
}

// Base64URL adds base64 url encoding of the value to the entry.
//
// gccgo's libgo is based on go1.18, which lacks base64.Encoding.AppendEncode,
// so the encoding is written into a scratch buffer first.
func (e *Entry) Base64URL(key string, value []byte) *Entry {
	if e == nil {
		return nil
	}

	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	buf := make([]byte, base64.URLEncoding.EncodedLen(len(value)))
	base64.URLEncoding.Encode(buf, value)
	e.buf = append(e.buf, buf...)
	e.buf = append(e.buf, '"')
	return e
}

package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

var (
	LineSeparator      = []byte("\r\n")
	ValueSeparator     = []byte(":")
	ErrMalformedHeader = fmt.Errorf("malformed header")
)

func NewHeaders() Headers {
	return Headers{}
}

func parseSingleHeader(fieldLine []byte) (string, string, error) {
	rKey, rValue, _ := bytes.Cut(fieldLine, ValueSeparator)
	key := bytes.TrimSpace(rKey)
	value := bytes.TrimSpace(rValue)
	return string(key), string(value), nil
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false

	for {
		ls := bytes.Index(data[read:], LineSeparator)
		if ls == -1 {
			break
		}
		if ls == 0 {
			done = true
			read += len(LineSeparator)
			break
		}
		if vi := bytes.Index(data, ValueSeparator); vi == -1 || (vi > 0 && data[vi-1] == ' ') {
			return 0, false, ErrMalformedHeader
		}

		key, value, err := parseSingleHeader(data[read : read+ls])
		if err != nil {
			return 0, false, ErrMalformedHeader
		}

		h[key] = value
		read += ls + len(LineSeparator)
	}

	return read, done, nil
}

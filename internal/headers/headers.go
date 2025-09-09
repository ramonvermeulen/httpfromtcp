package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

var (
	LineSeparator               = []byte("\r\n")
	ValueSeparator              = []byte(":")
	ErrMalformedHeaderFieldLine = fmt.Errorf("malformed header fieldLine")
	ErrMalformedHeaderFieldName = fmt.Errorf("malformed header fieldName")
)

func NewHeaders() Headers {
	return Headers{}
}

func isValidToken(bytes []byte) bool {
	for _, b := range bytes {
		switch {
		case b >= 'A' && b <= 'Z':
		case b >= 'a' && b <= 'z':
		case b >= '0' && b <= '9':
		case b == '!' || b == '#' || b == '$' || b == '%' || b == '&' || b == '\'' ||
			b == '*' || b == '+' || b == '-' || b == '.' || b == '^' || b == '_' ||
			b == '`' || b == '|' || b == '~':
		default:
			return false
		}
	}
	return true
}

func parseSingleHeader(fieldLine []byte) (string, string, error) {
	rKey, rValue, _ := bytes.Cut(fieldLine, ValueSeparator)
	key := bytes.ToLower(bytes.TrimSpace(rKey))
	if !isValidToken(key) || len(key) == 0 {
		return "", "", ErrMalformedHeaderFieldName
	}
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
			return 0, false, ErrMalformedHeaderFieldLine
		}

		key, value, err := parseSingleHeader(data[read : read+ls])
		if err != nil {
			return 0, false, ErrMalformedHeaderFieldLine
		}

		h[key] = value
		read += ls + len(LineSeparator)
	}

	return read, done, nil
}

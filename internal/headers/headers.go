package headers

import (
	"bytes"
	"fmt"
	"strings"
)

var (
	LineSeparator               = []byte("\r\n")
	ValueSeparator              = []byte(":")
	ErrMalformedHeaderFieldLine = fmt.Errorf("malformed header fieldLine")
	ErrMalformedHeaderFieldName = fmt.Errorf("malformed header fieldName")
)

func isValidToken(bytes []byte) bool {
	for _, b := range bytes {
		found := false
		if b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || b >= '0' && b <= '9' {
			found = true
		}

		switch b {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~', ':':
			found = true
		}
		if !found {
			return false
		}
	}

	return true
}

func parseSingleHeader(fieldLine []byte) (string, string, error) {
	rKey, rValue, _ := bytes.Cut(fieldLine, ValueSeparator)
	key := bytes.TrimSpace(bytes.ToLower(rKey))
	if !isValidToken(key) || len(key) == 0 {
		return "", "", ErrMalformedHeaderFieldName
	}
	value := bytes.TrimSpace(rValue)
	return string(key), string(value), nil
}

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Get(key string) (string, bool) {
	v, ok := h[strings.ToLower(key)]
	return v, ok
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)

	if v, ok := h[key]; ok {
		h[key] = fmt.Sprintf("%s, %s", v, value)
	} else {
		h[key] = value
	}
}

func (h Headers) Replace(key, value string) {
	name := strings.ToLower(key)
	h[name] = value
}

func (h Headers) Del(key string) {
	name := strings.ToLower(key)
	delete(h, name)
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
			return 0, false, fmt.Errorf("malformed header fieldLine: %w", err)
		}
		h.Set(key, value)

		read += ls + len(LineSeparator)
	}

	return read, done, nil
}

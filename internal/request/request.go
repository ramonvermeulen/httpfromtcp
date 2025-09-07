package request

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
}

var HTTPMethods = []string{
	"GET",
	"HEAD",
	"POST",
	"PUT",
	"DELETE",
	"CONNECT",
	"OPTIONS",
	"TRACE",
}

var (
	ERR_MALFORMED_REQUEST_LINE   = fmt.Errorf("malformed request line")
	ERR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported HTTP version")
	ERR_UNSUPOERTED_HTTP_METHOD  = fmt.Errorf("unsupported HTTP method")
)

func parseRequestLine(line string) (*RequestLine, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, ERR_MALFORMED_REQUEST_LINE
	}

	versionParts := strings.Split(parts[2], "/")
	requestTarget := parts[1]
	method := parts[0]

	if len(versionParts) != 2 || versionParts[0] != "HTTP" || versionParts[1] != "1.1" {
		return nil, ERR_UNSUPPORTED_HTTP_VERSION
	}

	match := slices.Contains(HTTPMethods, method)
	if !match {
		return nil, ERR_UNSUPOERTED_HTTP_METHOD
	}

	return &RequestLine{
		HTTPVersion:   versionParts[1],
		RequestTarget: requestTarget,
		Method:        method,
	}, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var requestLine *RequestLine
	for i, line := range strings.Split(string(data), "\r\n") {
		// for this exercise we only care about the request line
		if i > 0 {
			break
		}
		requestLine, err = parseRequestLine(line)
		if err != nil {
			return nil, err
		}
	}

	return &Request{
		RequestLine: *requestLine,
	}, nil
}

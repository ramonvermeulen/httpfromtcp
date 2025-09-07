package request

import (
	"bytes"
	"fmt"
	"io"
	"slices"
)

var HTTPMethods = [][]byte{
	[]byte("GET"),
	[]byte("HEAD"),
	[]byte("POST"),
	[]byte("PUT"),
	[]byte("DELETE"),
	[]byte("CONNECT"),
	[]byte("OPTIONS"),
	[]byte("TRACE"),
}

var (
	ErrMalformedRequestLine   = fmt.Errorf("malformed request line")
	ErrUnsupportedHttpVersion = fmt.Errorf("unsupported HTTP version")
	ErrUnsupportedHttpMethod  = fmt.Errorf("unsupported HTTP method")
	ErrParsingInDoneState     = fmt.Errorf("attempted to parse request in done state")
	LineSeparator             = []byte("\r\n")
)

type requestState int

const (
	Initialized requestState = iota
	Done                     = iota
)

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	state       requestState
}

func (r *Request) parse(data []byte) (int, error) {
outer:
	for {
		switch r.state {
		case Initialized:
			bytesProcessed, requestLine, err := parseRequestLine(data)
			if err != nil {
				return 0, err
			}
			if bytesProcessed == 0 {
				break outer
			}

			r.RequestLine = *requestLine
			r.state = Done
			return len(data), nil
		case Done:
			return 0, ErrParsingInDoneState
		}
	}
	return 0, nil
}

func parseRequestLine(line []byte) (int, *RequestLine, error) {
	if !bytes.Contains(line, LineSeparator) {
		return 0, nil, nil
	}

	line = bytes.SplitN(line, LineSeparator, 2)[0]
	parts := bytes.Split(line, []byte(" "))
	if len(parts) != 3 {
		return 0, nil, ErrMalformedRequestLine
	}

	versionParts := bytes.Split(parts[2], []byte("/"))
	requestTarget := parts[1]
	method := parts[0]

	if len(versionParts) != 2 || !bytes.Equal(versionParts[0], []byte("HTTP")) || !bytes.Equal(versionParts[1], []byte("1.1")) {
		return 0, nil, ErrUnsupportedHttpVersion
	}

	if !slices.ContainsFunc(HTTPMethods, func(m []byte) bool {
		return bytes.Equal(m, method)
	}) {
		return 0, nil, ErrUnsupportedHttpMethod
	}

	return len([]byte(line)), &RequestLine{
		HTTPVersion:   string(versionParts[1]),
		RequestTarget: string(requestTarget),
		Method:        string(method),
	}, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{}
	request.state = Initialized
	buffer := make([]byte, 1024)
	bufferLen := 0

	for request.state != Done {
		n, err := reader.Read(buffer[bufferLen:])
		if err != nil {
			break
		}
		bufferLen += n
		bytesProcessed, err := request.parse(buffer[:bufferLen])
		if err != nil {
			return nil, err
		}
		if bytesProcessed > 0 {
			buffer = buffer[bytesProcessed:]
		}
	}

	return request, nil
}

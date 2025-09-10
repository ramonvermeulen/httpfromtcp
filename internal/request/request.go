package request

import (
	"bytes"
	"fmt"
	"io"
	"slices"

	"github.com/ramonvermeulen/httpfromtcp/internal/headers"
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
	Initialized    requestState = iota
	ParsingHeaders requestState = iota
	Done                        = iota
)

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       requestState
}

func NewRequest() *Request {
	return &Request{
		Headers: headers.NewHeaders(),
		state:   Initialized,
	}
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for r.state != Done {
		currentData := data[read:]

		switch r.state {
		case Initialized:
			bytesProcessed, requestLine, err := parseRequestLine(currentData)
			if err != nil {
				return 0, err
			}
			if bytesProcessed == 0 {
				break outer
			}
			read += bytesProcessed

			r.RequestLine = *requestLine
			r.state = ParsingHeaders

		case ParsingHeaders:
			bytesProcessed, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}
			if bytesProcessed == 0 {
				break outer
			}
			read += bytesProcessed

			if done {
				r.state = Done
			}

		case Done:
			return 0, ErrParsingInDoneState
		}
	}
	return read, nil
}

func parseRequestLine(line []byte) (int, *RequestLine, error) {
	lsIdx := bytes.Index(line, LineSeparator)
	if lsIdx == -1 {
		return 0, nil, nil
	}

	line = line[:lsIdx]
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

	return len([]byte(line)) + len(LineSeparator), &RequestLine{
		HTTPVersion:   string(versionParts[1]),
		RequestTarget: string(requestTarget),
		Method:        string(method),
	}, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := NewRequest()
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
			bufferLen -= bytesProcessed
			buffer = buffer[bytesProcessed:]
		}
	}
	return request, nil
}

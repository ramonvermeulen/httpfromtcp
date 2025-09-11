package request

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strconv"

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
	ErrMalformedRequestLine    = fmt.Errorf("malformed request line")
	ErrMalformedContentLength  = fmt.Errorf("malformed request line")
	ErrUnsupportedHTTPVersion  = fmt.Errorf("unsupported HTTP version")
	ErrUnsupportedHTTPMethod   = fmt.Errorf("unsupported HTTP method")
	ErrParsingInDoneState      = fmt.Errorf("attempted to parse request in done state")
	ErrBodyExceedContentLength = fmt.Errorf("body exceeds content-length")
	ErrBodyWithinContentLength = fmt.Errorf("body exceeds content-length")
	LineSeparator              = []byte("\r\n")
)

type requestState int

const (
	StateInitialized requestState = iota
	StateHeaders     requestState = iota
	StateBody        requestState = iota
	StateDone                     = iota
)

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        string

	state requestState
}

func NewRequest() *Request {
	return &Request{
		Headers: headers.NewHeaders(),
		state:   StateInitialized,
	}
}

func getIntFromHeader(h headers.Headers, key string, defaultValue int) int {
	valueStr, exists := h.Get(key)
	if !exists {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func (r *Request) hasBody() bool {
	cl := getIntFromHeader(r.Headers, "content-length", 0)
	return cl != 0
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for r.state != StateDone {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}

		switch r.state {
		case StateInitialized:
			bp, rl, err := parseRequestLine(currentData)
			if err != nil {
				return 0, err
			}
			if bp == 0 {
				break outer
			}
			r.RequestLine = *rl
			r.state = StateHeaders
			read += bp

		case StateHeaders:
			bp, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}
			if bp == 0 {
				break outer
			}

			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}
			read += bp

		case StateBody:
			cl := getIntFromHeader(r.Headers, "content-length", 0)

			remaining := min(cl-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) > cl {
				return 0, ErrBodyExceedContentLength
			}

			if len(r.Body) == cl {
				r.state = StateDone
				break outer
			}

		case StateDone:
			return 0, ErrParsingInDoneState

		default:
			panic("This should never happen")
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
		return 0, nil, ErrUnsupportedHTTPVersion
	}

	if !slices.ContainsFunc(HTTPMethods, func(m []byte) bool {
		return bytes.Equal(m, method)
	}) {
		return 0, nil, ErrUnsupportedHTTPMethod
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

	for request.state != StateDone {
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

	if getIntFromHeader(request.Headers, "content-length", 0) != len(request.Body) {
		return nil, ErrBodyWithinContentLength
	}

	return request, nil
}

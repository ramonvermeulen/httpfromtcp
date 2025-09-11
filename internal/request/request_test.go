package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestFromReader(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)

	// Test: Invalid number of parts in request line
	reader = &chunkReader{
		data:            "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Standard Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	h, _ := r.Headers.Get("host")
	assert.Equal(t, "localhost:42069", h)
	h, _ = r.Headers.Get("user-agent")
	assert.Equal(t, "curl/7.81.0", h)
	h, _ = r.Headers.Get("accept")
	assert.Equal(t, "*/*", h)

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
	require.Nil(t, r)

	// Test: Empty Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Empty(t, r.Headers)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)

	// Test: Duplicate headers with allowed spacing and case insensitivity
	reader = &chunkReader{
		data:            "GET /cats HTTP/1.1\r\n   Foo:   bar\r\n    foo: BiZ\r\n\r\n",
		numBytesPerRead: 16,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, 1, len(r.Headers))
	h, _ = r.Headers.Get("foo")
	assert.Equal(t, "bar, BiZ", h)

	// Test: Missing end of headers
	reader = &chunkReader{
		data:            "GET /cats HTTP/1.1\r\nfoo: BiZ\r\n",
		numBytesPerRead: 16,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, 1, len(r.Headers))
	h, _ = r.Headers.Get("foo")
	assert.Equal(t, "BiZ", h)
}

func TestRequestBody(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 200\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Empty body with 0 reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"user-agent: some-user-agent/foo\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 13,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Empty(t, r.Body)
	assert.Equal(t, 0, len(r.Body))
	assert.Equal(t, "", string(r.Body))

	// Test: Empty body with no reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"user-agent: some-user-agent/foo\r\n" +
			"\r\n",
		numBytesPerRead: 13,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Empty(t, r.Body)
	assert.Equal(t, 0, len(r.Body))
	assert.Equal(t, "", string(r.Body))
}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

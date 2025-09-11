package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders_Parse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	h, _ := headers.Get("host")
	assert.Equal(t, "localhost:42069", h)
	assert.Equal(t, 25, n)
	assert.True(t, done)

	// Test: two times the same header, should append
	headers = NewHeaders()
	data = []byte("FooBar:   hello\r\nFooBar: noway\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	require.Len(t, headers, 1)
	h, _ = headers.Get("foobar")
	assert.Equal(t, "hello, noway", h)
	assert.Equal(t, 34, n)
	assert.True(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("   Host:    localhost:42069   \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	h, _ = headers.Get("host")
	assert.Equal(t, "localhost:42069", h)
	assert.Equal(t, 34, n)
	assert.True(t, done)

	// Test: valid two headers with existing header
	headers = NewHeaders()
	data = []byte("User-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	require.Len(t, headers, 2)
	h, _ = headers.Get("user-agent")
	require.Equal(t, "curl/7.81.0", h)
	h, _ = headers.Get("accept")
	require.Equal(t, "*/*", h)
	assert.Equal(t, 40, n)
	assert.True(t, done)

	// Test: valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	require.Empty(t, headers)
	require.Equal(t, 2, n)
	require.True(t, done)

	// Test: invalid special characters
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	require.Equal(t, 0, n)
	require.False(t, done)

	// Test: invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

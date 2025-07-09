package command

import (
	"io"
	"strings"
	// "strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data         string
	currentPos   int
	numBytesRead int
}

func (r *chunkReader) Read(p []byte) (n int, err error) {
	if r.currentPos >= len(r.data) {
		return 0, io.EOF
	}
	if r.currentPos+r.numBytesRead > len(r.data) {
		r.currentPos = len(r.data)
	}
	n = copy(p, r.data[r.currentPos:r.currentPos+r.numBytesRead])
	r.currentPos += n
	return n, nil
}

func TestCommandParse(t *testing.T) {
	// Test: Good ECHO Request with 2 params
	chunkReader := &chunkReader{
		data:         "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n",
		currentPos:   0,
		numBytesRead: 3,
	}
	r, err := CommandFromReader(chunkReader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, 2, r.NumberOfArgs)

	// Test: Bad ECHO Request with 2 params
	_, err = CommandFromReader(strings.NewReader("$2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"))
	require.Error(t, err)

}

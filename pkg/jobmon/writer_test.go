package jobmon

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaptureWriter(t *testing.T) {
	t.Run("simple-test", func(t *testing.T) {
		incomingStream := bytes.NewBufferString("")
		cw := newCaptureWriter(incomingStream, 100)

		_, _ = cw.Write([]byte(""))
		assert.Equal(t, "", cw.String())

		_, _ = cw.Write([]byte("0123456789"))
		assert.Equal(t, "0123456789", cw.String())

		_, _ = cw.Write([]byte(strings.Repeat("0123456789", 9)))
		assert.Equal(t, strings.Repeat("0123456789", 10), cw.String())
	})

	t.Run("full-copy", func(t *testing.T) {
		incomingStream := bytes.NewBufferString("")
		cw := newCaptureWriter(incomingStream, 100)

		_, _ = cw.Write([]byte(strings.Repeat("0123456789", 10)))
		assert.Equal(t, strings.Repeat("0123456789", 10), cw.String())
	})

	t.Run("overflow", func(t *testing.T) {
		incomingStream := bytes.NewBufferString("")
		cw := newCaptureWriter(incomingStream, 10)

		_, _ = cw.Write([]byte(strings.Repeat("0123456789", 2)))
		assert.Equal(t, strings.Repeat("0123456789", 1), cw.String())
	})

	t.Run("one-byte-overflow", func(t *testing.T) {
		incomingStream := bytes.NewBufferString("")
		cw := newCaptureWriter(incomingStream, 10)

		_, _ = cw.Write([]byte("01234567890"))
		assert.Equal(t, "1234567890", cw.String())
	})

	t.Run("one-byte-overflow-1", func(t *testing.T) {
		incomingStream := bytes.NewBufferString("")
		cw := newCaptureWriter(incomingStream, 10)

		_, _ = cw.Write([]byte(strings.Repeat("001234", 1)))
		_, _ = cw.Write([]byte(strings.Repeat("56789", 1)))
		assert.Equal(t, strings.Repeat("0123456789", 1), cw.String())
	})

	t.Run("one-byte-overflow-2", func(t *testing.T) {
		incomingStream := bytes.NewBufferString("")
		cw := newCaptureWriter(incomingStream, 10)

		_, _ = cw.Write([]byte(strings.Repeat("00123", 1)))
		_, _ = cw.Write([]byte(strings.Repeat("456789", 1)))
		assert.Equal(t, strings.Repeat("0123456789", 1), cw.String())
	})

	t.Run("overflow-many", func(t *testing.T) {
		incomingStream := bytes.NewBufferString("")
		cw := newCaptureWriter(incomingStream, 100)

		_, _ = cw.Write([]byte(strings.Repeat("0123456789", 20)))
		assert.Equal(t, strings.Repeat("0123456789", 10), cw.String())
	})
}

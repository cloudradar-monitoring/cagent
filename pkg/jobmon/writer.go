package jobmon

import (
	"io"
)

// captureWriter copies all data to destination writer and captures last N bytes
type captureWriter struct {
	buf []byte
	n   int
	dst io.Writer
}

func newCaptureWriter(dst io.Writer, n int) *captureWriter {
	return &captureWriter{buf: make([]byte, 0, n), n: n, dst: dst}
}

func (w *captureWriter) String() string {
	return string(w.buf)
}

func (w *captureWriter) Write(p []byte) (n int, err error) {
	gotLen := len(p)
	if gotLen >= w.n {
		w.buf = p[gotLen-w.n:]
	} else if gotLen > 0 {
		newLength := len(w.buf) + gotLen
		if newLength <= w.n {
			w.buf = append(w.buf, p...)
		} else {
			truncateIndex := newLength - w.n - 1
			w.buf = append(w.buf[truncateIndex-1:], p...)
		}
	}

	return w.dst.Write(p)
}

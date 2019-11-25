package common

import (
	"fmt"
	"time"
)

type MeasurementsMap map[string]interface{}

func (mm MeasurementsMap) AddWithPrefix(prefix string, m MeasurementsMap) MeasurementsMap {
	for k, v := range m {
		mm[prefix+k] = v
	}
	return mm
}

func (mm MeasurementsMap) AddInnerWithPrefix(prefix string, m MeasurementsMap) MeasurementsMap {
	mm[prefix] = m

	return mm
}

// Timestamp type allows marshaling time.Time struct as Unix timestamp value
type Timestamp time.Time

func (t *Timestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(*t).Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}

// LimitedBuffer allows to store last N bytes written to it, discarding unneeded bytes
type LimitedBuffer struct {
	buf []byte
	n   int
}

func NewLimitedBuffer(n int) *LimitedBuffer {
	return &LimitedBuffer{buf: make([]byte, 0, n), n: n}
}

func (w *LimitedBuffer) String() string {
	return string(w.buf)
}

func (w *LimitedBuffer) Write(p []byte) (n int, err error) {
	gotLen := len(p)
	if gotLen >= w.n {
		w.buf = p[gotLen-w.n-1:]
	} else if gotLen > 0 {
		newLength := len(w.buf) + gotLen
		if newLength <= w.n {
			w.buf = append(w.buf, p...)
		} else {
			truncateIndex := newLength - w.n - 1
			w.buf = append(w.buf[truncateIndex-1:], p...)
		}
	}

	return gotLen, nil
}

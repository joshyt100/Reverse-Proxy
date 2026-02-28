package proxy

import (
	"io"
	"sync/atomic"
)

type doneReadCloser struct {
	rc   io.ReadCloser
	done func()
	once uint32
}

func (d *doneReadCloser) Read(p []byte) (int, error) {
	return d.rc.Read(p)
}

func (d *doneReadCloser) Close() error {
	if d.done != nil && atomic.SwapUint32(&d.once, 1) == 0 {
		d.done()
	}
	return d.rc.Close()
}

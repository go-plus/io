package io

import (
	"context"
	"fmt"
	"go.uber.org/atomic"
	"io"
)

type multiReader struct {
	ctx        context.Context
	cancel     context.CancelFunc
	count      *atomic.Int32
	chanReader chan *chanReader
	readers    []io.Reader
}

type chanReader struct {
	index int
	p     []byte
	n     int
	err   error
}

func (m *multiReader) Read(p []byte) (n int, err error) {
	if m.chanReader == nil {
		if len(m.readers) == 1 {
			if r, ok := m.readers[0].(*multiReader); ok {
				m.readers = r.readers
			}
		}
		m.ctx, m.cancel = context.WithCancel(context.TODO())
		m.chanReader = make(chan *chanReader)
		m.count = atomic.NewInt32(int32(len(m.chanReader)))
		for i := 0; i < len(m.readers); i++ {
			go func(ctx context.Context, cb chan<- *chanReader, index int, reader io.Reader) {
				defer m.count.Dec()
				for {
					r := &chanReader{
						index: index,
						p:     make([]byte, len(p)),
					}
					r.n, r.err = reader.Read(r.p)
					select {
					case <-ctx.Done():
						return
					case cb <- r:
					}
					if r.err != nil {
						return
					}
				}
			}(m.ctx, m.chanReader, i, m.readers[i])
		}
	}

	for {
		if m.count.Load() <= 0 {
			break
		}
		select {
		case <-m.ctx.Done():
			break
		case r := <-m.chanReader:
			if r.err != nil && r.err != io.EOF {
				if m.cancel != nil {
					m.cancel()
					m.cancel = nil
				}
				return 0, fmt.Errorf("index(%d) was error:%w", r.index, r.err)
			} else if r.err == io.EOF {
				continue
			}
			n := copy(p, r.p[:r.n])
			return n, nil
		}
	}
	if m.chanReader != nil {
		close(m.chanReader)
		m.chanReader = nil
	}
	return 0, io.EOF
}

// MultiReader returns a Reader that's the logical concatenation of
// the provided input readers. They're read sequentially. Once all
// inputs have returned EOF, Read will return EOF.  If any of the readers
// return a non-nil, non-EOF error, Read will return that error.
func MultiReader(readers ...io.Reader) io.Reader {
	r := make([]io.Reader, len(readers))
	copy(r, readers)
	return &multiReader{
		readers: r,
	}
}

package io

import (
	"io"
)

type multiReader struct {
	count      int
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
		m.chanReader = make(chan *chanReader)
		for i := 0; i < len(m.readers); i++ {
			go func(cb chan<- *chanReader, index int, reader io.Reader) {
				for {
					r := &chanReader{
						index: index,
						p:     make([]byte, len(p)),
					}
					r.n, r.err = reader.Read(r.p)
					cb <- r
					if r.err != nil {
						return
					}
				}
			}(m.chanReader, i, m.readers[i])
		}
	}

	for {
		if m.count >= len(m.readers) {
			close(m.chanReader)
			m.chanReader = nil
			break
		}
		r := <-m.chanReader
		if r.err != nil && r.err != io.EOF {
			close(m.chanReader)
			m.chanReader = nil
			return 0, r.err
		} else {
			if r.err == io.EOF {
				m.count++
				continue
			}
			n := copy(p, r.p[:r.n])
			return n, nil
		}
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
	return &multiReader{readers: r}
}

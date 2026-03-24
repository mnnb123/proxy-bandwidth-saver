package meter

import (
	"io"
	"sync/atomic"
)

// CountingReader wraps an io.Reader and counts bytes read
type CountingReader struct {
	reader io.Reader
	count  atomic.Int64
}

func NewCountingReader(r io.Reader) *CountingReader {
	return &CountingReader{reader: r}
}

func (cr *CountingReader) Read(p []byte) (int, error) {
	n, err := cr.reader.Read(p)
	cr.count.Add(int64(n))
	return n, err
}

func (cr *CountingReader) BytesRead() int64 {
	return cr.count.Load()
}

// CountingWriter wraps an io.Writer and counts bytes written
type CountingWriter struct {
	writer io.Writer
	count  atomic.Int64
}

func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{writer: w}
}

func (cw *CountingWriter) Write(p []byte) (int, error) {
	n, err := cw.writer.Write(p)
	cw.count.Add(int64(n))
	return n, err
}

func (cw *CountingWriter) BytesWritten() int64 {
	return cw.count.Load()
}

// CountingReadCloser wraps an io.ReadCloser with byte counting
type CountingReadCloser struct {
	*CountingReader
	closer io.Closer
}

func NewCountingReadCloser(rc io.ReadCloser) *CountingReadCloser {
	return &CountingReadCloser{
		CountingReader: NewCountingReader(rc),
		closer:         rc,
	}
}

func (crc *CountingReadCloser) Close() error {
	return crc.closer.Close()
}

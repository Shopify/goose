package syncio

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

// Buffer is a wrapper on top of bytes.Buffer, but with a RWMutex to protect concurrent calls.
type Buffer struct {
	l sync.RWMutex
	b bytes.Buffer
}

func (rw *Buffer) Read(p []byte) (n int, err error) {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.Read(p)
}

func (rw *Buffer) Write(p []byte) (n int, err error) {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.Write(p)
}

func (rw *Buffer) ReadFrom(r io.Reader) (n int64, err error) {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.ReadFrom(r)
}

func (rw *Buffer) WriteTo(w io.Writer) (n int64, err error) {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.WriteTo(w)
}

func (rw *Buffer) ReadRune() (r rune, size int, err error) {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.ReadRune()
}

func (rw *Buffer) UnreadRune() error {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.UnreadRune()
}

func (rw *Buffer) WriteRune(r rune) (n int, err error) {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.WriteRune(r)
}

func (rw *Buffer) ReadByte() (byte, error) {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.ReadByte()
}

func (rw *Buffer) ReadBytes(delim byte) (line []byte, err error) {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.ReadBytes(delim)
}

func (rw *Buffer) UnreadByte() error {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.UnreadByte()
}

func (rw *Buffer) WriteByte(c byte) error {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.WriteByte(c)
}

func (rw *Buffer) ReadString(delim byte) (line string, err error) {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.ReadString(delim)
}

func (rw *Buffer) WriteString(s string) (n int, err error) {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.WriteString(s)
}

func (rw *Buffer) Next(n int) []byte {
	rw.l.Lock()
	defer rw.l.Unlock()
	return rw.b.Next(n)
}

func (rw *Buffer) Reset() {
	rw.l.Lock()
	rw.b.Reset()
	rw.l.Unlock()
}

func (rw *Buffer) Truncate(n int) {
	rw.l.Lock()
	rw.b.Truncate(n)
	rw.l.Unlock()
}

func (rw *Buffer) Grow(n int) {
	rw.l.Lock()
	rw.b.Grow(n)
	rw.l.Unlock()
}

func (rw *Buffer) String() string {
	if rw == nil {
		// Special case, useful in debugging.
		return "<nil>"
	}

	rw.l.RLock()
	defer rw.l.RUnlock()
	return rw.b.String()
}

func (rw *Buffer) Bytes() []byte {
	rw.l.RLock()
	defer rw.l.RUnlock()
	return rw.b.Bytes()
}

func (rw *Buffer) Len() int {
	rw.l.RLock()
	defer rw.l.RUnlock()
	return rw.b.Len()
}

func (rw *Buffer) Cap() int {
	rw.l.RLock()
	defer rw.l.RUnlock()
	return rw.b.Cap()
}

var _ interface {
	io.ReadWriter
	io.ReaderFrom
	io.WriterTo
	io.RuneScanner
	io.ByteScanner
	io.ByteWriter
	fmt.Stringer
} = &Buffer{}

func NewBuffer(buf []byte) *Buffer { return &Buffer{b: *bytes.NewBuffer(buf)} }

func NewBufferString(s string) *Buffer {
	return NewBuffer([]byte(s))
}

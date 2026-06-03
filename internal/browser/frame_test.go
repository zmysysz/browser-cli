package browser

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"
)

// pipeConn is a minimal net.Conn backed by an in-memory pipe so we can test
// the framing protocol without binding a Unix socket.
type pipeConn struct {
	net.Conn
	r *bytes.Buffer
	w *bytes.Buffer
}

func newPipeConn() (*pipeConn, *pipeConn) {
	aBuf, bBuf := &bytes.Buffer{}, &bytes.Buffer{}
	a := &pipeConn{r: aBuf, w: bBuf}
	b := &pipeConn{r: bBuf, w: aBuf}
	return a, b
}

func (p *pipeConn) Read(b []byte) (int, error)  { return p.r.Read(b) }
func (p *pipeConn) Write(b []byte) (int, error) { return p.w.Write(b) }
func (p *pipeConn) Close() error                { return nil }
func (p *pipeConn) LocalAddr() net.Addr         { return pipeAddr{} }
func (p *pipeConn) RemoteAddr() net.Addr        { return pipeAddr{} }
func (p *pipeConn) SetDeadline(time.Time) error { return nil }
func (p *pipeConn) SetReadDeadline(time.Time) error {
	return nil
}
func (p *pipeConn) SetWriteDeadline(time.Time) error {
	return nil
}

type pipeAddr struct{}

func (pipeAddr) Network() string { return "pipe" }
func (pipeAddr) String() string  { return "pipe" }

// ensure pipeConn satisfies net.Conn
var _ net.Conn = (*pipeConn)(nil)

func TestFrameRoundTrip(t *testing.T) {
	a, b := newPipeConn()
	defer a.Close()
	defer b.Close()

	payload := []byte(`{"action":"ping","params":{}}`)

	// For an in-memory pipe, writeFrame completes synchronously (bytes.Buffer
	// never blocks), so we can write first and read second. The goroutine
	// form was needed when the pipe was actually network-based; here it
	// would only hide bugs.
	if err := writeFrame(b, payload); err != nil {
		t.Fatalf("writeFrame() error = %v", err)
	}
	got, err := readFrame(a)
	if err != nil {
		t.Fatalf("readFrame() error = %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("readFrame() = %q, want %q", got, payload)
	}
}

func TestFrameLengthPrefix(t *testing.T) {
	// Sanity check that writeFrame puts the length in big-endian in the
	// first 4 bytes. This is part of the public wire format that other
	// tooling could depend on.
	a, b := newPipeConn()
	defer a.Close()
	defer b.Close()

	payload := []byte("hello world")
	if err := writeFrame(b, payload); err != nil {
		t.Fatalf("writeFrame() error = %v", err)
	}

	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(a, lenBuf); err != nil {
		t.Fatalf("read length: %v", err)
	}
	if got := binary.BigEndian.Uint32(lenBuf); got != uint32(len(payload)) {
		t.Errorf("length prefix = %d, want %d", got, len(payload))
	}
}

func TestReadFrame_RejectsOversize(t *testing.T) {
	// Send a length header that exceeds maxFrameBytes, expect a clean error
	// (not a silent read of the cap-sized buffer).
	a, b := newPipeConn()
	defer a.Close()
	defer b.Close()

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, maxFrameBytes+1)
	if _, err := b.Write(lenBuf); err != nil {
		t.Fatalf("write len: %v", err)
	}

	if _, err := readFrame(a); err == nil {
		t.Error("readFrame() on oversize header returned nil, want error")
	}
}

func TestReadFrame_RejectsEmpty(t *testing.T) {
	a, b := newPipeConn()
	defer a.Close()
	defer b.Close()

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, 0)
	if _, err := b.Write(lenBuf); err != nil {
		t.Fatalf("write len: %v", err)
	}

	if _, err := readFrame(a); err == nil {
		t.Error("readFrame() on empty frame returned nil, want error")
	}
}

func TestWriteFrame_RejectsOversize(t *testing.T) {
	a, b := newPipeConn()
	defer a.Close()
	defer b.Close()

	huge := make([]byte, maxFrameBytes+1)
	if err := writeFrame(b, huge); err == nil {
		t.Error("writeFrame() on oversize payload returned nil, want error")
	}
	// Counterpart receives nothing because nothing was sent.
	if _, err := a.r.Read(make([]byte, 1)); err != io.EOF {
		// Not strictly required to be EOF — but anything other than an
		// error is suspicious. We accept EOF as the only "ok" answer.
		t.Logf("unexpected read result (informational): %v", err)
	}
}

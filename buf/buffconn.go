package buf

import (
	"bufio"
	"net"
	"time"
)

type BufferedConn struct {
	conn   net.Conn
	reader *bufio.Reader
}

func NewBufferedConn(conn net.Conn, reader *bufio.Reader) *BufferedConn {
	return &BufferedConn{conn: conn, reader: reader}
}

func (b *BufferedConn) Read(p []byte) (int, error) {
	return b.reader.Read(p)
}
func (b *BufferedConn) Write(p []byte) (int, error) {
	return b.conn.Write(p)
}
func (b *BufferedConn) Close() error {
	return b.conn.Close()
}
func (b *BufferedConn) LocalAddr() net.Addr {
	return b.conn.LocalAddr()
}
func (b *BufferedConn) RemoteAddr() net.Addr {
	return b.conn.RemoteAddr()
}
func (b *BufferedConn) SetDeadline(t time.Time) error {
	return b.conn.SetDeadline(t)
}
func (b *BufferedConn) SetReadDeadline(t time.Time) error {
	return b.conn.SetReadDeadline(t)
}
func (b *BufferedConn) SetWriteDeadline(t time.Time) error {
	return b.conn.SetWriteDeadline(t)
}

package mpd

import (
	"net"
	"os"
	"time"

	"github.com/rotisserie/eris"
)

type dialer interface {
	Name() string
	Dial() (net.Conn, error)
}

var (
	defaultDialer = new(tcpDialer)
	// Useful for testing.
	testDialer = new(fakeDialer)
)

type tcpDialer struct{}

func (t *tcpDialer) Name() string {
	return "MPD dialer"
}

func (t *tcpDialer) Dial() (net.Conn, error) {
	host := os.Getenv("MPD_HOST")
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, eris.Wrapf(err, "can't dial %q", host)
	}
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   ips[0],
		Port: 6600,
	})
	if err != nil {
		return nil, eris.Wrap(err, "dial")
	}
	// Keep the TCP connection alive.
	conn.SetKeepAlive(true)
	return conn, nil
}

type fakeDialer struct{}

func (t *fakeDialer) Name() string {
	return "Test dialer"
}

func (t *fakeDialer) Dial() (net.Conn, error) {
	return &fakeConn{}, nil
}

type fakeConn struct{}

func (t *fakeConn) Read(b []byte) (int, error) {
	b = []byte("volume: 1\n")
	return len(b), nil
}

func (t *fakeConn) Write(b []byte) (int, error) {
	return len(b), nil
}

func (t *fakeConn) Close() error {
	return nil
}

func (t *fakeConn) LocalAddr() net.Addr {
	return nil
}

func (t *fakeConn) RemoteAddr() net.Addr {
	return nil
}

func (t *fakeConn) SetDeadline(p time.Time) error {
	return nil
}

func (t *fakeConn) SetReadDeadline(p time.Time) error {
	return nil
}

func (t *fakeConn) SetWriteDeadline(p time.Time) error {
	return nil
}

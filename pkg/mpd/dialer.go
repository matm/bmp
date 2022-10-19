package mpd

import (
	"net"
	"time"

	"github.com/rotisserie/eris"
)

type dialer interface {
	Name() string
	Dial(host string, port int) (net.Conn, error)
}

// DefaultPort is the default TCP port to the MPD service.
const DefaultPort = 6600

var (
	netDialer = new(tcpDialer)
	// Useful for testing.
	testDialer = new(fakeDialer)
)
var defaultDialer = netDialer

type tcpDialer struct{}

func (t *tcpDialer) Name() string {
	return "MPD dialer"
}

func (t *tcpDialer) Dial(host string, port int) (net.Conn, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, eris.Wrapf(err, "can't dial %q", host)
	}
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   ips[0],
		Port: port,
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

func (t *fakeDialer) Dial(host string, port int) (net.Conn, error) {
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

package coderws

import (
	"context"
	"math"
	"net"
	"net/http"

	"github.com/coder/websocket"
)

const MAX_WS_MESSAGE = math.MaxInt64 - 1

// NetConn wraps a coder websocket.Conn as a net.Conn
func NetConn(wsconn *websocket.Conn) net.Conn {
	return websocket.NetConn(context.Background(), wsconn, websocket.MessageBinary)
}

// ConnWithAddr attaches remote address info to a net.Conn
func ConnWithAddr(conn net.Conn, addr net.Addr) net.Conn {
	return &connAddr{conn, addr}
}

type connAddr struct {
	net.Conn
	net.Addr
}

func (ca *connAddr) RemoteAddr() net.Addr {
	return ca.Addr
}

// NewAddr creates a net.Addr from network and address string
func NewAddr(network, hostport string) net.Addr {
	return &addr{network, hostport}
}

type addr struct{ network, hostport string }

func (a *addr) Network() string { return a.network }
func (a *addr) String() string  { return a.hostport }

// Wrconn upgrades an HTTP connection to a WebSocket connection using coder/websocket
// and returns a net.Conn with remote address attached
func Wrconn(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	wsconn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}
	wsconn.SetReadLimit(MAX_WS_MESSAGE)
	conn := NetConn(wsconn)
	addr := NewAddr("websocket", r.RemoteAddr)
	conn = ConnWithAddr(conn, addr)
	return conn, nil
}

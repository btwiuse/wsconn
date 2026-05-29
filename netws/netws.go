package netws

import (
	"errors"
	"net"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

// NetConn returns a net.Conn from a golang.org/x/net/websocket.Conn.
// x/net/websocket.Conn already implements net.Conn directly.
// Sets PayloadType to BinaryFrame so outgoing frames use binary opcode.
func NetConn(c *websocket.Conn) net.Conn {
	c.PayloadType = websocket.BinaryFrame
	return c
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

// wsConn wraps a websocket.Conn with close notification so the server
// handler goroutine can exit cleanly.
type wsConn struct {
	*websocket.Conn
	once sync.Once
	done chan struct{}
}

func (c *wsConn) Close() error {
	err := c.Conn.Close()
	c.once.Do(func() { close(c.done) })
	return err
}

// Wrconn upgrades an HTTP connection to a WebSocket connection using
// x/net/websocket and returns a net.Conn.
func Wrconn(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	ch := make(chan *wsConn)

	// x/net/websocket.Handler's ServeHTTP does the upgrade via hijack
	// and calls the handler with the *websocket.Conn.
	// We run it in a goroutine and extract the conn through a channel.
	// The handler blocks until the caller closes the connection,
	// preventing the deferred rwc.Close() in serveWebSocket from firing.
	s := &websocket.Server{
		// nil Handshake skips origin checking, accepts non-browser clients
		Handler: func(ws *websocket.Conn) {
			ws.PayloadType = websocket.BinaryFrame
			c := &wsConn{Conn: ws, done: make(chan struct{})}
			ch <- c
			<-c.done
		},
	}

	go s.ServeHTTP(w, r)

	c := <-ch
	if c == nil {
		return nil, errors.New("netws: nil websocket connection")
	}
	return c, nil
}

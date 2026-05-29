package gorillaws

import (
	"io"
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const MAX_WS_MESSAGE = math.MaxInt64 - 1

/* http://www.gorillatoolkit.org/pkg/websocket
Connections support one concurrent reader and one concurrent writer.

Applications are responsible for ensuring that no more than one goroutine calls the write methods (NextWriter, SetWriteDeadline, WriteMessage, WriteJSON, EnableWriteCompression, SetCompressionLevel) concurrently and that no more than one goroutine calls the read methods (NextReader, SetReadDeadline, ReadMessage, ReadJSON, SetPongHandler, SetPingHandler) concurrently.
*/

// NetConn wraps a gorilla websocket.Conn as a net.Conn
func NetConn(c *websocket.Conn) net.Conn {
	return &netConn{
		Conn: c,
	}
}

// netConn makes a io.ReadWriteCloser from websocket.Conn, implementing the wetty.Master interface
// it is fed to wetty.New to create a WeTTY, bridging the websocket.Conn and local command
type netConn struct {
	*websocket.Conn
	mu sync.Mutex
}

func (wsw *netConn) Write(p []byte) (n int, err error) {
	wsw.mu.Lock()
	defer wsw.mu.Unlock()
	writer, err := wsw.Conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}
	defer writer.Close()
	return writer.Write(p)
}

func (wsw *netConn) Read(buf []byte) (int, error) {
	for {
		msgType, reader, err := wsw.Conn.NextReader()
		if err != nil {
			return 0, err
		}
		if msgType != websocket.BinaryMessage {
			continue
		}

		msg, err := io.ReadAll(reader)
		if err != nil {
			return 0, err
		}

		copy(buf, msg)

		n := len(msg)
		if n > len(buf) {
			n = len(buf)
		}

		return n, nil
	}
}

func (wsw *netConn) Close() error {
	return wsw.Conn.Close()
}

func (c *netConn) SetDeadline(t time.Time) (err error) {
	err = c.SetWriteDeadline(t)
	if err != nil {
		return
	}
	err = c.SetReadDeadline(t)
	return
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

var up = &websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Wrconn upgrades an HTTP connection to a WebSocket connection using gorilla/websocket
// and returns a net.Conn
func Wrconn(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	wsconn, err := up.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	wsconn.SetReadLimit(MAX_WS_MESSAGE)
	conn := NetConn(wsconn)
	return conn, nil
}

//go:build !gorilla

package wsconn

import (
	"math"
	"net"
	"net/http"

	"github.com/coder/websocket"
)

const MAX_WS_MESSAGE = math.MaxInt64 - 1 // -1 because the library adds a byte for the fin frame

func wrconn(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
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

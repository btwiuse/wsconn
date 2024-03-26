//go:build !gorilla

package wsconn

import (
	"net"
	"net/http"

	"nhooyr.io/websocket"
)

const MAX_WS_MESSAGE = 8 * 1024 * 1024 // 8MB

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

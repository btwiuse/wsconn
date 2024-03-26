//go:build !gorilla

package wsconn

import (
	"context"
	"net"

	"nhooyr.io/websocket"
)

func NetConn(wsconn *websocket.Conn) net.Conn {
	return websocket.NetConn(context.Background(), wsconn, websocket.MessageBinary)
}

package connection

import "github.com/gorilla/websocket"

type WebSocketConnection struct {
	*websocket.Conn
}

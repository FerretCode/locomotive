package subscribe

import (
	"context"
	"fmt"

	"github.com/coder/websocket"
)

// Railway tends to close the connection abruptly, this is needed to prevent any panics caused by reading from an abruptly closed connection
func SafeConnRead(conn *websocket.Conn, ctx context.Context) (mT websocket.MessageType, b []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from read panic: %v", r)
		}
	}()

	return conn.Read(ctx)
}

func SafeConnCloseNow(conn *websocket.Conn) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from close now panic: %v", r)
		}
	}()

	return conn.CloseNow()
}

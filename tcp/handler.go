package tcp

import (
	"context"
	"net"
)

// Handler TCP 连接的处理器
type Handler interface {
	Handle(connection net.Conn, ctx context.Context)
	Close() error
}

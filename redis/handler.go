package redis

import (
	"context"
	"io"
	"net"
	"simple_kvstorage/database"
	"simple_kvstorage/redis/resp/parser"
	"simple_kvstorage/redis/resp/reply"
	"simple_kvstorage/util/logger"
	"simple_kvstorage/util/sync/atomic"
	"strings"
	"sync"
)

// Handler the Redis handler
type Handler struct {
	// Redis 客户端连接的集合, map[Client]struct{}
	activeClient sync.Map
	// 存储引擎
	db database.Database
	// 当前 Handler 是否处于关闭过程中
	closing atomic.Boolean
}

func NewHandler(db database.Database) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Handle(connection net.Conn, ctx context.Context) {
	// 1. 如果处理器正在关闭中, 则不处理连接了
	if h.closing.Get() {
		_ = connection.Close()
	}

	// 2. 将连接封装进客户端中, 并记录客户端到活跃客户端的容器里
	client := NewDefaultClient(connection)
	h.activeClient.Store(client, struct{}{})

	// 3. 与客户端进行交互通信
	parseChan := parser.CreateParser(client.connection)
	for payload := range parseChan {
		// 返回给 Redis 客户端的回应
		var theReply reply.Reply

		if payload.Error != nil {
			// 发生 IO 错误, 表示 TCP 连接已经发生错误或已经关闭了
			if payload.Error == io.EOF || payload.Error == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Error.Error(), "use of closed network connection") {
				logger.Info("payload.Error 为 EOF.")
				h.closeClient(client)
				break
			}

			// 发生参数错误, 协议错误, 语法错误, 或其他错误. 则返回给客户端错误原因即可.
			errorReply, ok := payload.Error.(reply.ErrorReply)
			if !ok {
				errorReply = reply.NewStandardErrReply(payload.Error.Error())
			}

			theReply = errorReply
		} else {
			if payload.Data == nil {
				logger.Error("客户端的请求报文是空的.")
				continue
			}

			parsedReply, ok := payload.Data.(*reply.MultiBulkReply)
			if !ok {
				logger.Error("客户端的请求报文不是 Multi Bulk Reply.", string(payload.Data.ToBytes()))
				continue
			}

			// 接收到正常的命令报文, 执行命令
			dbReply := h.db.Exec(client, parsedReply.Args)
			if dbReply == nil {
				dbReply = reply.GetUnknownErrorReply()
			}

			theReply = dbReply
		}

		// 将 theReply 通过 TCP 写给客户端
		err := client.Write(theReply.ToBytes())
		if err != nil {
			h.closeClient(client)
			return
		}
	}
}

func (h *Handler) Close() error {
	logger.Info("Redis handler 即将关闭, 等待所有 Redis 连接的释放")
	h.closing.Set(true)
	h.activeClient.Range(func(key, value any) bool {
		_ = key.(*DefaultClient).Close()
		return true
	})
	h.db.Close()
	logger.Info("Redis handler 已关闭.")
	return nil
}

// closeClient 关闭一个 Redis 客户端连接
func (h *Handler) closeClient(client *DefaultClient) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeClient.Delete(client)
	logger.Info("Redis Client 已关闭", client.RemoteAddr())
}

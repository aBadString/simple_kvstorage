package core

import (
	"context"
	"io"
	"net"
	"runtime/debug"
	"simple_kvstorage/database"
	"simple_kvstorage/executor"
	"simple_kvstorage/persistent"
	"simple_kvstorage/resp"
	"simple_kvstorage/resp/reply"
	"simple_kvstorage/util/logger"
	"simple_kvstorage/util/sync/atomic"
	"strconv"
	"strings"
	"sync"
)

// Handler the Core handler
type Handler struct {
	// 客户端连接的集合, map[*Client]struct{}
	activeClient sync.Map
	// 当前 Handler 是否处于关闭过程中
	closing atomic.Boolean

	// 存储引擎
	dbs []database.DB
	// 持久化
	aof persistent.Persistent
}

func NewHandler(dbs []database.DB, aof persistent.Persistent) *Handler {
	return &Handler{dbs: dbs, aof: aof}
}

func (h *Handler) Handle(connection net.Conn, ctx context.Context) {
	// 1. 如果处理器正在关闭中, 则不处理连接了
	if h.closing.Get() {
		_ = connection.Close()
	}

	// 2. 将连接封装进客户端中, 并记录客户端到活跃客户端的容器里
	client := newClient(connection)
	h.activeClient.Store(client, struct{}{})

	// 3. 与客户端进行交互通信
	parseChan := resp.CreateParser(client.connection)
	for payload := range parseChan {
		// 给客户端的回应
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
				errorReply = reply.NewStandardErrorReply(payload.Error.Error())
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
			dbReply := h.Exec(client, parsedReply.Args)
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
	logger.Info("Handler 即将关闭, 等待所有 Client 连接的释放.")
	h.closing.Set(true)
	h.activeClient.Range(func(key, value any) bool {
		_ = key.(*Client).Close()
		return true
	})
	h.CloseDatabase()
	logger.Info("Handler 已关闭.")
	return nil
}

// closeClient 关闭一个客户端连接
func (h *Handler) closeClient(client *Client) {
	_ = client.Close()
	h.AfterClientClose(client)
	h.activeClient.Delete(client)
	logger.Info("Client 已关闭.")
}

// Exec 执行命令
func (h *Handler) Exec(client *Client, cmdLine executor.CmdLine) reply.Reply {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("recover 时发生错误.", err, '\n', string(debug.Stack()))
		}
	}()

	cmdName := strings.ToLower(string(cmdLine[0]))
	if cmdName == "select" {
		return h.execSelect(client, cmdLine)
	}

	// normal commands
	if h.aof != nil {
		h.aof.Persistence(client.GetDBIndex(), cmdLine)
	}
	selectedDB := h.dbs[client.GetDBIndex()]
	return executor.Exec(selectedDB, cmdLine)
}

// execSelect SELECT index
// 参考: https://redis.io/commands/select
func (h *Handler) execSelect(client *Client, cmdLine executor.CmdLine) reply.Reply {
	if len(cmdLine) != 2 {
		return reply.NewArgNumberErrorReply("select")
	}

	dbIndex, err := strconv.Atoi(string(cmdLine[1]))
	if err != nil {
		return reply.NewStandardErrorReply("ERROR invalid DB index")
	}
	if dbIndex >= len(h.dbs) {
		return reply.NewStandardErrorReply("ERROR DB index is out of range")
	}

	client.SelectDB(dbIndex)
	return reply.GetOkReply()
}

// AfterClientClose 一个客户端断开连接之后的清理工作
func (h *Handler) AfterClientClose(client *Client) {}

// CloseDatabase 关闭数据库
func (h *Handler) CloseDatabase() {}

package tcp

import (
	"bufio"
	"context"
	"io"
	"net"
	"simple_kvstorage/util/logger"
	"simple_kvstorage/util/sync/atomic"
	"simple_kvstorage/util/sync/wait"
	"sync"
	"testing"
	"time"
)

func TestListenAndServe(t *testing.T) {
	err := ListenAndServe(&Config{Address: "127.0.0.1:6379"}, &echoHandler{})
	if err != nil {
		t.Fatal(err)
		return
	}
}

type client struct {
	connection net.Conn
	waiting    wait.Wait
}

func (c *client) Close() error {
	c.waiting.WaitWithTimeout(10 * time.Second)
	return c.connection.Close()
}

type echoHandler struct {
	activeConnection sync.Map
	closing          atomic.Boolean
}

func (h *echoHandler) Handle(connection net.Conn, ctx context.Context) {
	// 1. 如果处理器正在关闭中, 则不处理连接了
	if h.closing.Get() {
		_ = connection.Close()
	}

	// 2. 将连接封装进客户端中, 并记录客户端到活跃客户端的容器里
	client := &client{connection: connection}
	h.activeConnection.Store(client, struct{}{})

	// 3. 与客户端进行交互通信
	reader := bufio.NewReader(client.connection)
	for {
		// 3.1. 接收客户端请求. 当客户端没有请求数据来的时候, 会阻塞在此处
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("连接已关闭.", client.connection.RemoteAddr())
				h.activeConnection.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}

		// 3.2. 执行业务逻辑, 并返回给客户端响应
		client.waiting.Add(1)
		b := []byte(msg)
		_, _ = client.connection.Write(b)
		client.waiting.Done()
	}
}

func (h *echoHandler) Close() error {
	h.closing.Set(true)
	h.activeConnection.Range(func(key, value any) bool {
		_ = key.(*client).Close()
		return true
	})
	return nil
}

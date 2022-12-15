package redis

import (
	"net"
	"simple_kvstorage/util/sync/wait"
	"sync"
	"time"
)

type DefaultClient struct {
	// TCP 的连接
	connection net.Conn
	// 当 Redis 前客户端连接的数据库序号
	selectedDB int

	waitingReply wait.Wait
	locker       sync.Mutex
}

func NewDefaultClient(connection net.Conn) *DefaultClient {
	return &DefaultClient{connection: connection}
}

func (c *DefaultClient) Write(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}

	c.locker.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.locker.Unlock()
	}()

	_, err := c.connection.Write(bytes)
	return err
}

func (c *DefaultClient) GetDBIndex() int {
	return c.selectedDB
}

func (c *DefaultClient) SelectDB(index int) {
	c.selectedDB = index
}

func (c *DefaultClient) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.connection.Close()
	return nil
}

func (c *DefaultClient) RemoteAddr() net.Addr {
	return c.connection.RemoteAddr()
}

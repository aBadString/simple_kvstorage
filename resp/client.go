package resp

import (
	"net"
	"simple_kvstorage/util/sync/wait"
	"sync"
	"time"
)

type defaultClient struct {
	// TCP 的连接
	connection net.Conn
	// 当前客户端连接的数据库序号
	selectedDB int

	waitingReply wait.Wait
	locker       sync.Mutex
}

func newDefaultClient(connection net.Conn) *defaultClient {
	return &defaultClient{connection: connection}
}

func (c *defaultClient) Write(bytes []byte) error {
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

func (c *defaultClient) GetDBIndex() int {
	return c.selectedDB
}

func (c *defaultClient) SelectDB(index int) {
	c.selectedDB = index
}

func (c *defaultClient) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.connection.Close()
	return nil
}

func (c *defaultClient) RemoteAddr() net.Addr {
	return c.connection.RemoteAddr()
}

package core

import (
	"net"
	"simple_kvstorage/util/sync/wait"
	"sync"
	"time"
)

// Client 描述了对客户端连接的操作
type Client struct {
	// TCP 的连接
	connection net.Conn
	// 当前客户端连接的数据库序号
	selectedDB int

	waitingReply wait.Wait
	locker       sync.Mutex
}

func newClient(connection net.Conn) *Client {
	return &Client{connection: connection}
}

// Write 向客户端写数据
func (c *Client) Write(bytes []byte) error {
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

// GetDBIndex 获取此客户端连接的数据库序号
func (c *Client) GetDBIndex() int {
	return c.selectedDB
}

// SelectDB 切换此客户端的数据库
// index 数据库序号
func (c *Client) SelectDB(index int) {
	c.selectedDB = index
}

func (c *Client) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.connection.Close()
	return nil
}

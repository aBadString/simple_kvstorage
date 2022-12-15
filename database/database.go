package database

import (
	"simple_kvstorage/common"
	"simple_kvstorage/redis/resp/reply"
)

// Database 存储引擎的接口
type Database interface {
	// Exec 执行命令
	Exec(client common.RedisClient, args [][]byte) reply.Reply

	// AfterClientClose 一个客户端断开连接之后的清理工作
	AfterClientClose(client common.RedisClient)
	// Close 关闭数据库
	Close()
}

// CmdLine 表示一行解析好的 Redis 命令
type CmdLine = [][]byte

// DataEntity 存储层的数据结构, 包括 string, list, hash, set 等
type DataEntity struct {
	Data interface{}
}
package common

// Client 描述了 Redis 客户端连接的操作
type Client interface {
	// Write 向 Redis 客户端写数据
	Write([]byte) error

	// GetDBIndex 获取此客户端连接的数据库序号
	GetDBIndex() int

	// SelectDB 切换此客户端的数据库
	// index 数据库序号
	SelectDB(index int)
}

package database

// DB 数据存储层
type DB interface {
	// Get 按 key 获取 val
	Get(key string) (val *DataEntity, exists bool)

	// Put 存入一个键值对
	Put(key string, val *DataEntity) (result int)

	// PutIfExists 若 key 在 Dict 里才存入这个键值对
	PutIfExists(key string, val *DataEntity) (result int)

	// PutIfAbsent 若 key 不在 Dict 里则存入这个键值对
	PutIfAbsent(key string, val *DataEntity) (result int)

	// Remove 按 key 删除一个键值对
	Remove(key string)

	// Removes 批量删除 Keys
	// 返回删除的键值对的数量
	Removes(keys ...string) int

	// Size 返回 Map 中键值对的数量
	Size() int

	// ForEach 遍历 Map, 对每个键值对应用 traverser 函数
	// traverser 用于遍历 Map 的函数, 当其返回 false 时停止继续遍历
	ForEach(traverser func(key string, val *DataEntity) bool)

	// Keys 获取全部的 key
	Keys() []string

	// RandomKeys 随机获取 limit 个 key
	RandomKeys(limit int) []string

	// RandomDistinctKeys 随机获取 limit 个不重复的 key
	RandomDistinctKeys(limit int) []string

	// Flush 清空数据库
	Flush()
}

// DataEntity 存储层的数据结构, 包括 string, list, hash, set 等
type DataEntity struct {
	Data interface{}
}

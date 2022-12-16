package datastructure

// Traverser 用于遍历字典 (Dict) 的函数, 当其返回 false 时停止继续遍历
type Traverser func(key string, val interface{}) bool

// Dict 键值对 (K-V) 数据结构的接口
type Dict interface {
	// Get 按 key 获取 val
	Get(key string) (val interface{}, exists bool)

	// Len 返回 Dict 中键值对的数量
	Len() int

	// Put 存入一个键值对
	Put(key string, val interface{}) (result int)

	// PutIfAbsent 若 key 不在 Dict 里则存入这个键值对
	PutIfAbsent(key string, val interface{}) (result int)

	// PutIfExists 若 key 在 Dict 里才存入这个键值对
	PutIfExists(key string, val interface{}) (result int)

	// Remove 按 key 删除一个键值对
	Remove(key string) (result int)

	// ForEach 遍历 Dict, 对每个键值对应用 Traverser 函数
	ForEach(traverser Traverser)

	// Keys 获取全部的 key
	Keys() []string

	// RandomKeys 随机获取 limit 个 key
	RandomKeys(limit int) []string

	// RandomDistinctKeys 随机获取 limit 个不重复的 key
	RandomDistinctKeys(limit int) []string

	// Clear 清空 Dict
	Clear()
}

package command

import (
	"simple_kvstorage/database"
	"simple_kvstorage/executor"
	"simple_kvstorage/resp/reply"
	"simple_kvstorage/util/wildcard"
)

func init() {
	executor.RegisterCommand(del, execDel, -2)
	executor.RegisterCommand(exists, execExists, -2)
	executor.RegisterCommand(keys, execKeys, 2)
	executor.RegisterCommand(flushDB, execFlushDB, -1)
	executor.RegisterCommand(_type, execType, 2)
	executor.RegisterCommand(rename, execRename, 3)
	executor.RegisterCommand(renameNx, execRenameNx, 3)
}

// execDel DEL key [key ...]
// 参考: https://redis.io/commands/del
func execDel(db database.DB, args [][]byte) reply.Reply {
	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}

	deleted := db.Removes(keys...)
	return reply.NewIntReply(int64(deleted))
}

// execExists EXISTS key [key ...]
// 参考: https://redis.io/commands/exists
func execExists(db database.DB, args [][]byte) reply.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		_, exists := db.Get(key)
		if exists {
			result++
		}
	}
	return reply.NewIntReply(result)
}

// execKeys KEYS pattern
// 参考: https://redis.io/commands/keys
func execKeys(db database.DB, args [][]byte) reply.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))
	result := make([][]byte, 0)
	db.ForEach(func(key string, _ *database.DataEntity) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.NewMultiBulkReply(result)
}

func execFlushDB(db database.DB, _ [][]byte) reply.Reply {
	db.Flush()
	return reply.GetOkReply()
}

// execType TYPE key
// 参考: https://redis.io/commands/type
func execType(db database.DB, args [][]byte) reply.Reply {
	key := string(args[0])
	entity, exists := db.Get(key)
	if !exists {
		return reply.NewStatusReply("none")
	}

	switch entity.Data.(type) {
	case []byte:
		return reply.NewStatusReply("string")
	case int:
		return reply.NewStatusReply("integer")
	}
	return reply.GetUnknownErrorReply()
}

// execRename RENAME key newkey
// 参考: https://redis.io/commands/rename
func execRename(db database.DB, args [][]byte) reply.Reply {
	key := string(args[0])
	newKey := string(args[1])

	entity, exists := db.Get(key)
	if !exists {
		return reply.NewStandardErrReply("no such key '" + key + "'")
	}

	db.Put(newKey, entity)
	db.Remove(key)
	return reply.GetOkReply()
}

// execRenameNx RENAMENX key newkey
// 参考 https://redis.io/commands/renamenx
func execRenameNx(db database.DB, args [][]byte) reply.Reply {
	key := string(args[0])
	newKey := string(args[1])

	_, exist := db.Get(newKey)
	if exist {
		return reply.NewIntReply(0)
	}

	entity, exist := db.Get(key)
	if !exist {
		return reply.NewStandardErrReply("no such key '" + key + "'")
	}

	db.Put(newKey, entity)
	db.Removes(key)
	return reply.NewIntReply(1)
}

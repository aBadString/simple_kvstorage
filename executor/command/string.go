package command

import (
	"simple_kvstorage/database"
	"simple_kvstorage/executor"
	"simple_kvstorage/resp/reply"
)

func init() {
	executor.RegisterCommand(get, execGet, 2)
	executor.RegisterCommand(set, execSet, -3)
	executor.RegisterCommand(setNx, execSetNX, 3)
	executor.RegisterCommand(getSet, execGetSet, 3)
	executor.RegisterCommand(strLen, execStrLen, 2)
}

// execGet GET key
// 参考: https://redis.io/commands/get
func execGet(db database.DB, args [][]byte) reply.Reply {
	key := string(args[0])
	entity, exist := db.Get(key)
	if !exist {
		return reply.GetNullBulkReply()
	}

	bytes, exist := entity.Data.([]byte)
	if !exist {
		return reply.GetWrongTypeErrorReply()
	}

	return reply.NewBulkReply(bytes)
}

// execSet SET key value
// 参考: https://redis.io/commands/set
func execSet(db database.DB, args [][]byte) reply.Reply {
	key := string(args[0])
	value := args[1]
	db.Put(key, &database.DataEntity{Data: value})
	return reply.GetOkReply()
}

// execSetNX SETNX key value
// 参考: https://redis.io/commands/setnx
func execSetNX(db database.DB, args [][]byte) reply.Reply {
	key := string(args[0])
	value := args[1]
	result := db.PutIfAbsent(key, &database.DataEntity{Data: value})
	return reply.NewIntReply(int64(result))
}

// execGetSet GETSET key value
// 参考: https://redis.io/commands/getset
func execGetSet(db database.DB, args [][]byte) reply.Reply {
	key := string(args[0])
	value := args[1]

	entity, exists := db.Get(key)
	db.Put(key, &database.DataEntity{Data: value})

	if !exists {
		return reply.GetNullBulkReply()
	}

	old, exist := entity.Data.([]byte)
	if !exist {
		return reply.GetWrongTypeErrorReply()
	}

	return reply.NewBulkReply(old)
}

// execStrLen STRLEN key
// 参考 https://redis.io/commands/strlen
func execStrLen(db database.DB, args [][]byte) reply.Reply {
	key := string(args[0])
	entity, exists := db.Get(key)
	if !exists {
		return reply.GetNullBulkReply()
	}

	bytes, exist := entity.Data.([]byte)
	if !exist {
		return reply.GetWrongTypeErrorReply()
	}

	l := len(bytes)
	return reply.NewIntReply(int64(l))
}

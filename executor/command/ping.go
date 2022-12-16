package command

import (
	"simple_kvstorage/database"
	"simple_kvstorage/executor"
	"simple_kvstorage/resp/reply"
)

func init() {
	executor.RegisterCommand(ping, execPing, -1)
}

// execPing PING [message]
// 参考: https://redis.io/commands/ping
var execPing executor.CommandExecutor = func(db database.DB, argsWithoutCmdName [][]byte) reply.Reply {
	switch len(argsWithoutCmdName) {
	case 0:
		return &reply.PongReply{}
	case 1:
		return reply.NewStatusReply(string(argsWithoutCmdName[0]))
	default:
		return reply.NewArgNumberErrorReply(ping)
	}
}

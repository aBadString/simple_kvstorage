package executor

import (
	"simple_kvstorage/database"
	"simple_kvstorage/resp/reply"
	"strings"
)

// CmdLine 表示一行解析好的命令.
// 例如: set key value
type CmdLine = [][]byte

// Exec executes command within one database
func Exec(db database.DB, cmdLine CmdLine) reply.Reply {
	cmdName := strings.ToLower(string(cmdLine[0]))

	cmd, exist := cmdTable[cmdName]
	if !exist {
		return reply.NewStandardErrReply("ERROR unknown command '" + cmdName + "'")
	}
	if !cmd.validateArity(cmdLine) {
		return reply.NewArgNumberErrorReply(cmdName)
	}

	return cmd.executor(db, cmdLine[1:])
}

/* --- command --- */

// cmdTable 记录了支持的全部 Redis 命令
var cmdTable = make(map[string]*command)

// RegisterCommand 注册一个命令
func RegisterCommand(cmdName string, executor CommandExecutor, arity int) {
	cmdName = strings.ToLower(cmdName)
	cmdTable[cmdName] = &command{
		executor: executor,
		arity:    arity,
	}
}

// CommandExecutor 命令所对应的要执行的函数
// argsWithoutCmdName 是不包括命令名称的, 即 argsWithoutCmdName = cmdLine[1:]
type CommandExecutor func(db database.DB, argsWithoutCmdName [][]byte) reply.Reply

// command 描述了一个 Redis 命令
type command struct {
	executor CommandExecutor

	// arity 命令所需要的参数个数 (包括 cmdName).
	// 当 arity >= 0 时, 此命令的参数数量必须要刚好是 arity.
	// 当 arity  < 0 时, 此命令的参数数量需要大于 arity 的绝对值.
	// 详见 validateArity 函数.
	// for example: the arity of `get` is 2, `mget` is -2
	arity int
}

// validateArity 校验命令参数的数量是否正确
func (c *command) validateArity(cmdLine CmdLine) bool {
	argNum := len(cmdLine)
	if c.arity >= 0 {
		return argNum == c.arity
	}
	return argNum >= -c.arity
}

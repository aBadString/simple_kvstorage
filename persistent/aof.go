package persistent

import (
	"os"
	"simple_kvstorage/executor"
	"simple_kvstorage/resp/reply"
	"simple_kvstorage/util/logger"
	"strconv"
	"strings"
)

type Persistent interface {
	Persistence(dbIndex int, cmdLine executor.CmdLine)
}

// cmdPersistent 记录了需要持久化的命令
var cmdPersistent map[string]interface{}

func init() {
	var cmd = []string{
		"del",
		"flushDB",
		"rename",
		"renameNx",
		"set",
		"setNx",
		"getSet",
	}

	cmdPersistent = make(map[string]interface{})
	for i := range cmd {
		cmdName := strings.ToLower(cmd[i])
		cmdPersistent[cmdName] = struct{}{}
	}
}

type AofPersistent struct {
	// 是否开启持久化
	enable bool
	// 持久化文件
	aofFile *os.File
	// 当前数据库序号
	currentDB int

	aofChan chan *aofCmd
}

func NewAofPersistent(aofFilename string, enable bool) *AofPersistent {
	p := &AofPersistent{enable: enable}

	var err error
	p.aofFile, err = os.OpenFile(aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil
	}

	// 开启一个持久化协程
	p.aofChan = make(chan *aofCmd, 1<<8)
	go p.persistenceFromChan()
	return p
}

type aofCmd struct {
	dbIndex int
	cmdLine executor.CmdLine
}

// Persistence 持久化刚刚执行成功的命令
func (p *AofPersistent) Persistence(dbIndex int, cmdLine executor.CmdLine) {
	if p.enable && p.aofChan != nil {

		// 判断命令是否需要持久化
		cmdName := strings.ToLower(string(cmdLine[0]))
		_, exist := cmdPersistent[cmdName]
		if !exist {
			return
		}

		p.aofChan <- &aofCmd{
			dbIndex: dbIndex,
			cmdLine: cmdLine,
		}
	}
}

func (p *AofPersistent) persistenceFromChan() {
	// 先写一条 select 0 命令
	p.currentDB = -1

	for cmd := range p.aofChan {
		// 数据库切换了则写入一条 select db 命令
		if p.currentDB != cmd.dbIndex {
			data := reply.NewMultiBulkReply(toCmdLine("select", strconv.Itoa(cmd.dbIndex))).ToBytes()
			_, err := p.aofFile.Write(data)
			if err != nil {
				logger.Warn(err)
				continue
			}
			p.currentDB = cmd.dbIndex
		}

		data := reply.NewMultiBulkReply(cmd.cmdLine).ToBytes()
		_, err := p.aofFile.Write(data)
		if err != nil {
			logger.Warn(err)
		}
	}
}

func toCmdLine(cmd ...string) [][]byte {
	args := make([][]byte, len(cmd))
	for i, s := range cmd {
		args[i] = []byte(s)
	}
	return args
}

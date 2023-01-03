package persistent

import (
	"io"
	"os"
)

// LoadAof 加载持久化数据
// aofFilename 持久化数据文件
// handle 函数用于处理数据加载逻辑, 参数 connection 是一个流, 可以从中不断获取命令
func LoadAof(aofFilename string, handle func(connection io.ReadWriteCloser)) {
	handle(newLoadAofConn(aofFilename))
}

type loadAofConn struct {
	aofFile *os.File
}

func newLoadAofConn(aofFilename string) *loadAofConn {
	file, err := os.Open(aofFilename)
	if err != nil {
		return nil
	}
	return &loadAofConn{aofFile: file}
}

func (l *loadAofConn) Read(b []byte) (n int, err error) {
	return l.aofFile.Read(b)
}

func (l *loadAofConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (l *loadAofConn) Close() error {
	return l.aofFile.Close()
}

package persistent

import (
	"net"
	"os"
	"time"
)

type LoadAofConn struct {
	aofFile *os.File
}

func NewLoadAofConn(aofFilename string) *LoadAofConn {
	file, err := os.Open(aofFilename)
	if err != nil {
		return nil
	}
	return &LoadAofConn{aofFile: file}
}

func (l *LoadAofConn) Read(b []byte) (n int, err error) {
	return l.aofFile.Read(b)
}

func (l *LoadAofConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (l *LoadAofConn) Close() error {
	return l.aofFile.Close()
}

func (l *LoadAofConn) LocalAddr() net.Addr {
	panic("implement me")
}

func (l *LoadAofConn) RemoteAddr() net.Addr {
	panic("implement me")
}

func (l *LoadAofConn) SetDeadline(t time.Time) error {
	panic("implement me")
}

func (l *LoadAofConn) SetReadDeadline(t time.Time) error {
	panic("implement me")
}

func (l *LoadAofConn) SetWriteDeadline(t time.Time) error {
	panic("implement me")
}

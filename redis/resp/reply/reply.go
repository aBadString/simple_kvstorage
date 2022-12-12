package reply

import (
	"bytes"
	"strconv"
)

// Reply 序列化回复给客户端的消息为字节数组
type Reply interface {
	ToBytes() []byte
}

// BulkReply 回复一个字符串
type BulkReply struct {
	s []byte
}

func (r *BulkReply) ToBytes() []byte {
	if len(r.s) == 0 {
		return nullBulk
	}

	return []byte("$" + strconv.Itoa(len(r.s)) + CRLF + string(r.s) + CRLF)
}

func NewBulkReply(s []byte) *BulkReply {
	return &BulkReply{s: s}
}

type MultiBulkReply struct {
	ms [][]byte
}

func (r *MultiBulkReply) ToBytes() []byte {
	var bytesBuffer bytes.Buffer
	var bulkReply BulkReply

	bytesBuffer.WriteString("*" + strconv.Itoa(len(r.ms)) + CRLF)
	for _, s := range r.ms {
		bulkReply.s = s
		bytesBuffer.Write(bulkReply.ToBytes())
	}

	return bytesBuffer.Bytes()
}

func NewMultiBulkReply(ms [][]byte) *MultiBulkReply {
	return &MultiBulkReply{ms: ms}
}

// StatusReply 回复状态
type StatusReply struct {
	Status string
}

func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

func NewStatusReply(status string) *StatusReply {
	return &StatusReply{Status: status}
}

// IntReply 回复一个 64 位的整数
type IntReply struct {
	Code int64
}

func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

func NewIntReply(code int64) *IntReply {
	return &IntReply{Code: code}
}

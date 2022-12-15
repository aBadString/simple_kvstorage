package reply

const (
	CRLF = "\r\n"
)

// 一些固定的回复字符串常量
var (
	pong           = []byte("+PONG" + CRLF)
	ok             = []byte("+OK" + CRLF)
	nullBulk       = []byte("$-1" + CRLF)
	emptyMultiBulk = []byte("*0" + CRLF)
	no             = []byte("")
)

// 一些单例的 Reply 对象
var (
	pongReply           = &PongReply{}
	okReply             = &OkReply{}
	nullBulkReply       = &NullBulkReply{}
	emptyMultiBulkReply = &EmptyMultiBulkReply{}
	noReply             = &NoReply{}
)

// PongReply 回复 pong, 对于客户端的 ping
type PongReply struct {
}

func (*PongReply) ToBytes() []byte {
	return pong
}

// GetPongReply 获取一个 PongReply 对象 (全局单例的)
func GetPongReply() *PongReply {
	return pongReply
}

// OkReply 回复 ok
type OkReply struct {
}

func (*OkReply) ToBytes() []byte {
	return ok
}

// GetOkReply 获取一个 OkReply 对象 (全局单例的)
func GetOkReply() *OkReply {
	return okReply
}

// NullBulkReply 回复 nullBulk
type NullBulkReply struct {
}

func (*NullBulkReply) ToBytes() []byte {
	return nullBulk
}

// GetNullBulkReply 获取一个 NullBulkReply 对象 (全局单例的)
func GetNullBulkReply() *NullBulkReply {
	return nullBulkReply
}

// EmptyMultiBulkReply 回复 emptyMultiBulk
type EmptyMultiBulkReply struct {
}

func (*EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulk
}

// GetEmptyMultiBulkReply 获取一个 EmptyMultiBulkReply 对象 (全局单例的)
func GetEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return emptyMultiBulkReply
}

// NoReply 没有回复
type NoReply struct {
}

func (*NoReply) ToBytes() []byte {
	return no
}

// GetNoReply 获取一个 NoReply 对象 (全局单例的)
func GetNoReply() *NoReply {
	return noReply
}

package reply

type ErrorReply interface {
	error
	Reply
}

// IsErrorReply 返回 true, 当 reply 是一个错误回复时
func IsErrorReply(reply Reply) bool {
	return reply.ToBytes()[0] == '-'
}

// 一些单例的 ErrorReply 对象
var (
	unknownErrorReply   = &UnknownErrorReply{}
	syntaxErrorReply    = &SyntaxErrorReply{}
	wrongTypeErrorReply = &WrongTypeErrorReply{}
)

// UnknownErrorReply 未知错误
type UnknownErrorReply struct {
}

func (*UnknownErrorReply) Error() string {
	return "Error unknown"
}

func (r *UnknownErrorReply) ToBytes() []byte {
	return []byte("-" + r.Error() + CRLF)
}

func GetUnknownErrorReply() *UnknownErrorReply {
	return unknownErrorReply
}

// ArgNumberErrorReply 参数数量不正确
type ArgNumberErrorReply struct {
	Cmd string
}

func (r *ArgNumberErrorReply) Error() string {
	return "Error wrong number of arguments for '" + r.Cmd + "' command"
}

func (r *ArgNumberErrorReply) ToBytes() []byte {
	return []byte("-" + r.Error() + CRLF)
}

func NewArgNumberErrorReply(cmd string) *ArgNumberErrorReply {
	return &ArgNumberErrorReply{
		Cmd: cmd,
	}
}

// SyntaxErrorReply 表示客户端发来的命令存在语法错误
type SyntaxErrorReply struct {
}

func (*SyntaxErrorReply) Error() string {
	return "Error syntax error"
}

func (r *SyntaxErrorReply) ToBytes() []byte {
	return []byte("-" + r.Error() + CRLF)
}

func GetSyntaxErrReply() *SyntaxErrorReply {
	return syntaxErrorReply
}

// WrongTypeErrorReply 数据类型错误
type WrongTypeErrorReply struct {
}

func (*WrongTypeErrorReply) Error() string {
	return "WRONG_TYPE Operation against a key holding the wrong kind of value"
}

func (r *WrongTypeErrorReply) ToBytes() []byte {
	return []byte("-" + r.Error() + CRLF)
}

func GetWrongTypeErrorReply() *WrongTypeErrorReply {
	return wrongTypeErrorReply
}

// ProtocolErrorReply 协议错误, 客户端发来的指令不符合 RESP 规范
type ProtocolErrorReply struct {
	Msg string
}

func (r *ProtocolErrorReply) Error() string {
	return "ERROR Protocol error: '" + r.Msg + "'"
}

func (r *ProtocolErrorReply) ToBytes() []byte {
	return []byte("-" + r.Error() + CRLF)
}

func NewProtocolErrorReply(msg string) *ProtocolErrorReply {
	return &ProtocolErrorReply{
		Msg: msg,
	}
}

// StandardErrorReply 标准错误回复
type StandardErrorReply struct {
	Message string
}

func (r *StandardErrorReply) Error() string {
	return r.Message
}

func (r *StandardErrorReply) ToBytes() []byte {
	return []byte("-" + r.Message + CRLF)
}

func NewStandardErrorReply(message string) *StandardErrorReply {
	return &StandardErrorReply{Message: message}
}

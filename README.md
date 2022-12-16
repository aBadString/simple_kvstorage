# 1. 基础工具封装

- `config/config.go`: 用于读取配置文件
- `util/logger/files.go`: 封装了文件操作函数
- `util/logger/logger.go`: 封装了日志记录函数
- `util/sync/atomic/bool.go`: Boolean, 原子的 bool 类型
- `util/sync/wait/wait.go`: Wait, 带有的超时机制的 sync.WaitGroup


# 2. TCP 服务器

1. 服务器进程启动时, 开启一个 Socket 监听端口, 即 `Listener`.
2. 这个监听器工作在主协程上, 接着主协程一直循环, 并阻塞式等待客户端的连接, 即 `Accept`.
3. 主协程没接受到一个新的客户端连接, 就开启一个协程来处理.
4. 当服务器进程将要退出之前, 主协程将等待正在处理连接的其他协程全部结束后, 才退出.  
   收到外部的进程退出信号时, 则不等待其他协程退出, 直接关闭服务器.


# 3. RESP 协议

> [RESP protocol spec | Redis](https://redis.io/topics/protocol)

**回复类型**:  
1. 正常回复 (单行字符串 Simple String): 以 `+` 开头, `\r\n` 结尾的字符串. `"+OK\r\n"`
2. 错误回复: 以 `-` 开头, `\r\n` 结尾的字符串. `"-Error message\r\n"`
3. 整数: 以 `:` 开头, `\r\n` 结尾的字符串, 组成字符串的每个字符都是数字字符. `:1024\r\n`
4. 多行字符串 (Bulk): 以 `$` 开头, 后跟实际字节长度, 再跟`\r\n`, 随后是字符串, 最后以 `\r\n` 结尾. `"$5\r\nhello\r\n"`
5. 数组 (Multi Bulk): 以 `*` 开头, 后跟元素数量

> In RESP, the first byte determines the data type:  
> - For Simple Strings, the first byte of the reply is "+"
> - For Errors, the first byte of the reply is "-"
> - For Integers, the first byte of the reply is ":"
> - For Bulk Strings, the first byte of the reply is "$"
> - For Arrays, the first byte of the reply is "*"

例子: 

- 空字符串表示: `"$0\r\n\r\n"`
- 空数组: `"*0\r\n"`
- `nil` 的表示: `"$-1\r\n"`, `"*-1\r\n"`
- 数组 `["hello",nil,"world"]` 的表示如下:  
    ```
    *3\r\n
    $5\r\n
    hello\r\n
    $-1\r\n
    $5\r\n
    world\r\n
    ```

---

使用 Go 对每种 Reply 的描述一般由三个部分组成:  
1. Reply 的类型定义.
2. 对 `Reply` 接口的实现.
3. 用于获取该 Reply 类型对象的工厂方法, 创建新对象的方法名为 `NewXxxReply`, 获取全局单例对象方法名为 `GetXxxReply`.

一个固定回复 `"+PONG\r\n"` 的类型:  
```go
// 固定的回复字符串常量
var pong = []byte("+PONG" + CRLF)

// 单例的 Reply 对象
var pongReply = &PongReply{}


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
```

错误的回复类型:  
```go
type ErrorReply interface {
    error
    Reply
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
```


# 4. 命令

> [Commands | Redis](https://redis.io/commands)
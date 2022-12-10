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
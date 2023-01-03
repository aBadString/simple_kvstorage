package tcp

import (
	"context"
	"io"
	"net"
	"os"
	"os/signal"
	"simple_kvstorage/util/logger"
	"sync"
	"syscall"
)

// Handler TCP 连接的处理器
type Handler interface {
	Handle(connection io.ReadWriteCloser, ctx context.Context)
	Close() error
}

// Config TCP 服务器的配置
type Config struct {
	// Address 监听地址, 形如 "127.0.0.1:6379"
	Address string
}

func ListenAndServe(config *Config, handler Handler) error {
	// 开启一个协程来等待进程关闭信号, 当收到进程关闭信号时会发消息到 closeChan
	closeChan := registerCloseServerSignal()

	// 1. 开启 TCP 监听
	listener, err := net.Listen("tcp", config.Address)
	if err != nil {
		logger.Error("服务器创建失败.", err)
		return err
	}
	logger.Info("服务器启动成功.", listener.Addr())
	var waiter sync.WaitGroup

	// 2. 注册服务器关闭处理
	closeServer := func() {
		_ = listener.Close()
		_ = handler.Close()
		logger.Info("服务器已关闭.")
	}
	// 2.1 主协程正常执行到 return 退出
	defer func() {
		logger.Info("主协程正常退出, 服务器即将关闭, 等待所有客户端连接的释放.")
		// 等待所有正在处理客户端连接的协程结束
		waiter.Wait()
		closeServer()
	}()
	// 2.2 收到进程关闭信号
	go func() {
		<-closeChan
		logger.Info("收到进程关闭信号, 服务器即将关闭, 直接丢弃所有客户端连接.")
		closeServer()
	}()

	ctx := context.Background()
	for {
		// 3. 等待客户机连接, 侦听并接受到此套接字的连接, 此方法在连接传入之前一直阻塞.
		connection, err := listener.Accept()
		if err != nil {
			logger.Error("接受连接时发生错误.", err)
			break
		}
		logger.Info("客户端已连接.", connection.RemoteAddr())

		waiter.Add(1)
		// 4. 一个协程负责处理一个连接
		go func() {
			defer func() {
				logger.Info("客户端断开连接.", connection.RemoteAddr())
				_ = connection.Close()
				waiter.Done()
			}()

			handler.Handle(connection, ctx)
		}()
	}
	return nil
}

// registerCloseSignal 注册进程关闭信号
func registerCloseServerSignal() <-chan struct{} {
	closeChan := make(chan struct{})
	signChan := make(chan os.Signal)
	signal.Notify(signChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sign := <-signChan
		// 当接收到关闭进程信号时从此处开始执行
		switch sign {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	return closeChan
}

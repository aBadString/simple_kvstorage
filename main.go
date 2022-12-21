package main

import (
	"context"
	"fmt"
	"os"
	"simple_kvstorage/config"
	"simple_kvstorage/core"
	"simple_kvstorage/database"
	_ "simple_kvstorage/executor/command" // For execute the `init` function
	"simple_kvstorage/persistent"
	"simple_kvstorage/tcp"
	"simple_kvstorage/util/logger"
)

func main() {
	// 0. 设置日志记录
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "redis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})

	// 1. 读取配置文件或使用默认配置
	const configFile string = "redis.conf"
	if fileInfo, err := os.Stat(configFile); err == nil && !fileInfo.IsDir() {
		config.SetupConfig(configFile)
	} else {
		config.Properties = &config.ServerProperties{
			Bind:       "0.0.0.0",
			Port:       6379,
			Databases:  16,
			AppendOnly: false,
		}
	}

	tcpConfig := &tcp.Config{
		Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
	}

	// 2.1. 创建数据存储引擎
	dbs := make([]database.DB, config.Properties.Databases)
	for i := range dbs {
		dbs[i] = database.NewMapDB(i)
	}

	// 2.2. 创建持久化引擎
	var aofPersistent persistent.Persistent
	if config.Properties.AppendOnly {
		aofPersistent = persistent.NewAofPersistent(config.Properties.AppendFilename, true)
	}

	// 2.3. 加载持久化的数据
	loadAof(dbs)

	// 3. 启动 TCP 服务
	coreHandler := core.NewHandler(dbs, aofPersistent)
	err := tcp.ListenAndServe(tcpConfig, coreHandler)
	if err != nil {
		logger.Error(err)
		return
	}
}

func loadAof(dbs []database.DB) {
	handler := core.NewHandler(dbs, nil)
	handler.Handle(persistent.NewLoadAofConn(config.Properties.AppendFilename), context.Background())
}

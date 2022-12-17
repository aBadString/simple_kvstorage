package main

import (
	"fmt"
	"simple_kvstorage/config"
	"simple_kvstorage/core"
	"simple_kvstorage/database"
	_ "simple_kvstorage/executor/command" // For execute the `init` function
	"simple_kvstorage/tcp"
	"simple_kvstorage/util/logger"
)

func main() {
	tcpConfig := &tcp.Config{
		Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
	}

	dbs := make([]database.DB, config.Properties.Databases)
	for i := range dbs {
		dbs[i] = database.NewMapDB(i)
	}

	coreHandler := core.NewHandler(dbs)

	err := tcp.ListenAndServe(tcpConfig, coreHandler)
	if err != nil {
		logger.Error(err)
		return
	}
}

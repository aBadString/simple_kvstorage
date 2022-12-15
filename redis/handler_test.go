package redis

import (
	"simple_kvstorage/common"
	"simple_kvstorage/redis/resp/reply"
	"simple_kvstorage/tcp"
	"simple_kvstorage/util/logger"
	"testing"
)

func TestHandler(t *testing.T) {
	redisHandler := NewHandler(&echoDatabase{})
	err := tcp.ListenAndServe(&tcp.Config{Address: "127.0.0.1:6379"}, redisHandler)
	if err != nil {
		t.Fatal(err)
		return
	}
}

type echoDatabase struct {
}

func (d *echoDatabase) Exec(client common.RedisClient, args [][]byte) reply.Reply {
	logger.Debug("echoDatabase Exec")
	return reply.NewMultiBulkReply(args)
}

func (d *echoDatabase) AfterClientClose(client common.RedisClient) {
	logger.Debug("echoDatabase AfterClientClose")
}

func (d *echoDatabase) Close() {
	logger.Debug("echoDatabase Close")
}

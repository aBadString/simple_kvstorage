package parser

import (
	"bufio"
	"io"
	"runtime/debug"
	"simple_kvstorage/redis/resp/reply"
	"simple_kvstorage/util/logger"
	"strconv"
)

// Payload 解析器的输出
type Payload struct {
	Data  reply.Reply
	Error error
}

type parseState struct {
	// 当前解析的报文的类型 `+ - * $ :`
	msgType byte

	// readingMultiLine 当前解析的报文是否为多行的, 即 Bulk 或 Multi Bulk
	readingMultiLine bool

	// Bulk 中字节的数量
	bulkLen int64

	// expectedArgsCount 预期的参数数量, 即当前指令应该需要的参数的数量
	// 等于 MultiBulk 中 Bulk 的数量
	expectedArgsCount int

	// 已经解析好的部分
	args [][]byte
}

// finished 是否解析完毕
func (s *parseState) finished() bool {
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

// CreateParser 创建一个 RESP 协议的解析器, 其工作在另一个协程上.
// 不断将 reader 中的字节流解析为 Payload 放入管道 parseChan 中.
// 直到遇到 EOF 或其他 IO 错误, 才退出.
func CreateParser(reader io.Reader) <-chan *Payload {
	parseChan := make(chan *Payload)
	go func() {
		logger.Info("RESP 解析器开始工作.")
		defer func() {
			logger.Info("RESP 解析器停止工作.")
			close(parseChan)
			if err := recover(); err != nil {
				logger.Error("recover 时发生错误.", string(debug.Stack()))
			}
		}()
		parseToChan(reader, parseChan)
	}()
	return parseChan
}

func parseToChan(reader io.Reader, parseChan chan<- *Payload) {
	// 客户端连接在, 这个流就一直在
	var bufferReader = bufio.NewReader(reader)

	// 每一轮 for 循环, 调用一次 parse0 函数, 解析一个完整的命令请求报文
	for continueParsing := true; continueParsing; {
		// 1. 解析一个完整的报文
		theReply, err := parse0(
			func(state *parseState) ([]byte, error) {
				line, err, isIOError := readLine(bufferReader, state)
				if isIOError {
					// 发生 IO 错误, 则需要关闭这个客户端了
					continueParsing = false
				}
				return line, err
			},
		)

		// 2. 解析出错则将错误放入管道
		if err != nil {
			parseChan <- &Payload{Error: err}
			continue
		}

		// 3. 将解析正确的结果放入管道
		parseChan <- &Payload{Data: theReply}
	}
}

// parse0 解析一次完整的命令请求报文
// nextLine func(*parseState) ([]byte, error) 用于获取下一行报文
// 当 parse0 函数返回时, 表示其读取到了一个完整的报文
func parse0(nextLine func(*parseState) ([]byte, error)) (reply.Reply, error) {
	// 记录当前命令的解析状态
	var state = parseState{}

	//  每一轮 for 循环解析一行报文
	for {
		line, err := nextLine(&state)
		if err != nil {
			return nil, err
		}

		if !state.readingMultiLine {
			switch getType(line) {
			case '*': // Multi Bulk
				err := parseMultiBulkHeader(line, &state)
				if err != nil {
					return nil, err
				}
				if state.expectedArgsCount == 0 {
					return reply.GetEmptyMultiBulkReply(), nil
				}
				// continue parse Multi Bulk Body
			case '$': // Bulk
				err := parseBulkHeader(line, &state)
				if err != nil {
					return nil, err
				}
				if state.bulkLen == -1 {
					return reply.GetNullBulkReply(), nil
				}
				// continue parse Bulk Body
			default:
				single, err := parserSingle(line)
				return single, err
			}
		} else {
			err := readBody(line, &state)
			if err != nil {
				return nil, err
			}

			if state.finished() {
				switch state.msgType {
				case '*':
					return reply.NewMultiBulkReply(state.args), nil
				case '$':
					return reply.NewBulkReply(state.args[0]), nil
				}
			}

			// continue parse
		}
	}
}

// readLine 读取一行字符串. 可能是 Header (Multi Bulk Header, Bulk Header), Simple String 或 Bulk without Header.
//
// 分两种情况, 若不是多行字符串 (Bulk without Header), 则按 CRLF 为结尾划分; 若当前是多行字符串, 则读取给定的字节数.
// 返回 一行报文的字节数组, 是否发生 IO 异常, 具体的异常.
func readLine(bufferReader *bufio.Reader, state *parseState) ([]byte, error, bool) {
	var line []byte
	var err error

	// 1. 根据情况来读取一行字符串
	if state.bulkLen == 0 {
		// Simple String: +......CRLF
		line, err = bufferReader.ReadBytes('\n')
	} else if state.bulkLen > 0 {
		// Bulk: $字节长度CRLF......CRLF
		line = make([]byte, state.bulkLen+2) // 为 CRLF 预留两字节空间
		_, err = io.ReadFull(bufferReader, line)
		state.bulkLen = 0
	}

	if err != nil {
		return nil, err, true
	}

	// 2. 若所读到的一行字符串不是以 CRLF 结尾, 则表示客户端发送的数据不符合协议
	if !isEndWithCRLF(line) {
		return nil, reply.NewProtocolErrorReply("没有读到 CRLF"), false
	}

	return line, nil, false
}

// parseMultiBulkHeader 解析 MultiBulk 的首部
// MultiBulk: *元素个数CRLF   (Multi Bulk Header)
//
//	Bulk CRLF
//	Bulk CRLF
//	......
func parseMultiBulkHeader(multiBulk []byte, state *parseState) error {
	if len(multiBulk) < 4 || getType(multiBulk) != '*' || !isEndWithCRLF(multiBulk) {
		return reply.NewProtocolErrorReply(string(multiBulk))
	}

	// 读取 MultiBulk 中的元素个数, 即后续 Bulk 的数量
	expectedBulkCount, err := strconv.ParseInt(string(multiBulk[1:len(multiBulk)-2]), 10, 32)
	if err != nil || expectedBulkCount < -1 {
		return reply.NewProtocolErrorReply(string(multiBulk))
	}

	switch {
	case expectedBulkCount == -1: // *-1CRLF, 此处对 RESP 协议的实现将不识别 Null Array.
		return reply.NewProtocolErrorReply(string(multiBulk))
	case expectedBulkCount == 0: // *0CRLF 表示空数组 []
		state.expectedArgsCount = 0
	case expectedBulkCount > 0:
		state.msgType = getType(multiBulk)
		state.readingMultiLine = true // ?
		state.expectedArgsCount = int(expectedBulkCount)
		state.args = make([][]byte, 0, expectedBulkCount)
	}
	return nil
}

// parseBulkHeader 解析 Bulk 的首部
// Bulk: $字节长度CRLF   (Bulk Header)
//
//	String CRLF
//	String CRLF
//	......
func parseBulkHeader(bulk []byte, state *parseState) error {
	if len(bulk) < 4 || getType(bulk) != '$' || !isEndWithCRLF(bulk) {
		return reply.NewProtocolErrorReply(string(bulk))
	}

	// 读取 Bulk 中的的字节长度, 即后续 String 的长度
	expectedStringLength, err := strconv.ParseInt(string(bulk[1:len(bulk)-2]), 10, 64)
	if err != nil || expectedStringLength < -1 {
		return reply.NewProtocolErrorReply(string(bulk))
	}

	switch {
	case expectedStringLength == -1: // $-1CRLF 表示 nil
		state.bulkLen = -1
	case expectedStringLength == 0: // $0CRLF 表示空字符串 "", 此处对 RESP 协议的实现将不识别空字符串.
		return reply.NewProtocolErrorReply(string(bulk))
	default: // expectedStringLength > 0
		state.bulkLen = expectedStringLength
		state.msgType = getType(bulk)
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
	}
	return nil
}

// parserSingle 解析简单的单行报文, 包括 "+OK\r\n", "-Error message\r\n", ":1024\r\n"
func parserSingle(single []byte) (reply.Reply, error) {
	if len(single) < 3 || !isEndWithCRLF(single) {
		return nil, reply.NewProtocolErrorReply(string(single))
	}

	msg := string(single[1 : len(single)-2])

	switch getType(single) {
	case '+':
		return reply.NewStatusReply(msg), nil
	case '-':
		return reply.NewStandardErrReply(msg), nil
	case ':':
		i, err := strconv.ParseInt(msg, 10, 64)
		if err != nil {
			return nil, reply.NewProtocolErrorReply(string(single))
		}
		return reply.NewIntReply(i), nil
	default:
		return nil, reply.NewProtocolErrorReply(string(single))
	}
}

// readBody 读取 Multi Bulk 或 Bulk 的剩余部分
func readBody(body []byte, state *parseState) error {
	if getType(body) == '$' {
		// 在 Multi Bulk Header 之后调用了 readBody

		if len(body) < 4 || !isEndWithCRLF(body) {
			return reply.NewProtocolErrorReply(string(body))
		}

		expectedStringLength, err := strconv.ParseInt(string(body[1:len(body)-2]), 10, 64)
		if err != nil || expectedStringLength < -1 {
			return reply.NewProtocolErrorReply(string(body))
		}

		switch {
		case expectedStringLength <= 0: // expectedStringLength = -1 || 0
			state.bulkLen = 0
			state.args = append(state.args, []byte{})
		default: // expectedStringLength > 0
			state.bulkLen = expectedStringLength
		}
	} else {
		// 在 Bulk Header 之后调用了 readBody. 不会出现 getType(body) 为 * + - : 的情况

		state.args = append(state.args, body[0:len(body)-2])
	}
	return nil
}

func isEndWithCRLF(s []byte) bool {
	l := len(s)
	return l >= 2 && s[l-2] == '\r' && s[l-1] == '\n'
}

// getType 获取当前行的类型, `+ - * $ :`
func getType(line []byte) byte {
	return line[0]
}

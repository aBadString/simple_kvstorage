package parser

import (
	"io"
	"strings"
	"testing"
)

func TestParse0(t *testing.T) {
	var testCases = []string{
		"",
		"$-1",
		"$0",
		"*0",
		"+OK",
		"+PONG",
		"$5",
		"hello",
		":1024",
		"-Error message",
		"*3",
		"$5",
		"hello",
		"$-1",
		"$5",
		"world",
	}

	i := -1
	nextLine := func(*parseState) ([]byte, error) {
		i++
		if i < len(testCases) {
			return []byte(testCases[i] + "\r\n"), nil
		} else {
			return nil, io.EOF
		}
	}

	for i < len(testCases) {
		reply, err := parse0(nextLine)
		if err != nil {
			t.Log(err)
		} else {
			t.Log(reply)
		}
	}
}

func TestCreateParser(t *testing.T) {
	var testCases = []string{
		"\r\n",
		"$-1\r\n",
		"$0\r\n",
		"*0\r\n",
		"+OK\r\n",
		"+PONG\r\n",
		"+\r\n",
		"$5\r\nhello\r\n",
		":1024\r\n",
		"-Error message\r\n",
		"-\r\n",
		"*3\r\n$5\r\nhello\r\n$-1\r\n$5\r\nworld\r\n",
	}

	builder := strings.Builder{}
	for _, testCase := range testCases {
		builder.WriteString(testCase)
	}
	payloadChan := CreateParser(strings.NewReader(builder.String()))

	payloads := make([]*Payload, 0, len(testCases))
	for payload := range payloadChan {
		payloads = append(payloads, payload)

		if payload.Error == nil {
			t.Log(string(payload.Data.ToBytes()))
		} else {
			t.Log(payload.Error)
		}
	}

	t.Log(payloads)
}

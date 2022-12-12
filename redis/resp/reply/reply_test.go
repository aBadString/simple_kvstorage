package reply

import (
	"testing"
)

func TestReply(t *testing.T) {
	b1 := []byte("TEST1\nerer\r\nTEST")
	b2 := []byte("TEST2\nerer\r\nTEST\n")
	b3 := []byte("TEST3\nerer\r\nTEST\r\n")

	replies := []Reply{
		NewBulkReply(b1),
		NewMultiBulkReply([][]byte{
			b1, b2, b3,
		}),
		NewIntReply(1024),
		NewStatusReply("hello, RESP"),
	}

	for _, reply := range replies {
		t.Log(string(reply.ToBytes()))
	}
}
func TestFixedReply(t *testing.T) {
	fixedReplies := []Reply{
		GetPongReply(),
		GetOkReply(),
		GetNullBulkReply(),
		GetEmptyMultiBulkReply(),
		GetNoReply(),
	}

	for _, fixedReply := range fixedReplies {
		t.Log(string(fixedReply.ToBytes()))
	}
}
func TestErrorReply(t *testing.T) {
	errorReplies := []ErrorReply{
		GetSyntaxErrReply(),
		GetUnknownErrorReply(),
		GetWrongTypeErrorReply(),
		NewProtocolErrorReply("TEST"),
		NewArgNumberErrorReply("TEST"),
		NewStandardErrReply("TEST"),
	}

	for _, errorReply := range errorReplies {
		errorString := errorReply.Error()
		replyString := string(errorReply.ToBytes())

		if replyString == "-"+errorString+CRLF {
			t.Log("测试成功", errorString, replyString)
		} else {
			t.Error("测试失败", errorString, replyString)
		}
	}
}

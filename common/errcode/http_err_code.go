package errcode

import "fmt"

type HttpErrCode struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (h HttpErrCode) HttpBody() []byte {
	data := fmt.Sprintf("{code:%d,msg:%s}", h.Code, h.Msg)
	return []byte(data)
}

func MakeHttpErrCodeRespond(code int, msg string) []byte {
	return HttpErrCode{Code: code, Msg: msg}.HttpBody()
}

package protocol

import "sync"

// 游戏协议头（固定 6 字节）
// [4字节长度][2字节协议ID]

const (
	HeaderSize     = 4 + 4 + 4 // 长度固定
	lengthSizeTail = 4
	seqSizeTail    = 8
	msgIdSizeTail  = 12
)

var msgBufPool = sync.Pool{
	New: func() interface{} {
		// 这里按你的业务设置一个合理的初始大小，比如 4KB
		// 实际会根据需要扩容，但池里只会保存固定大小的对象
		return make([]byte, 4096)
	},
}

var headerBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, HeaderSize)
	},
}

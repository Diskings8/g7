package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
)

// gzipCompress 压缩数据
func gzipCompress(data []byte) []byte {
	var buf bytes.Buffer
	// 压缩级别 4 平衡速度和大小
	writer, _ := gzip.NewWriterLevel(&buf, 4)
	writer.Write(data)
	writer.Close()
	return buf.Bytes()
}

// gzipDecompress 解压数据
func gzipDecompress(data []byte) []byte {
	if len(data) == 0 {
		return make([]byte, 0)
	}
	reader, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
		return make([]byte, 0)
	}
	defer reader.Close()

	result, _ := io.ReadAll(reader)
	return result
}

func MarshalAndCompress(data any) []byte {
	if data == nil {
		return make([]byte, 0)
	}
	tempD, _ := json.Marshal(data)
	return gzipCompress(tempD)
}

func UnCompressAndUnmarshal(data []byte, tar any) {
	unCompressDAta := gzipDecompress(data)
	err := json.Unmarshal(unCompressDAta, tar)
	if err != nil {
		fmt.Println(err)
	}
	return
}

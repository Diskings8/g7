package createPlayer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"g7/common/protos/pb"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"
)

var globalClient = &http.Client{
	Transport: &http.Transport{
		MaxConnsPerHost:     100, // 关键！最大并发连接
		MaxIdleConns:        100, // 关键！
		MaxIdleConnsPerHost: 100, // 关键！
		IdleConnTimeout:     60 * time.Second,
		DisableKeepAlives:   false, // 必须开启长连接
	},
	Timeout: 20 * time.Second,
}

var GSuccess, GFail int32

func OneConnect(id int) {
	// 1. 构造要发送的数据
	reqData := &pb.Req_Http_CreatePlayer{
		UserID:   2044258565992091648,
		ServerID: 91001,
		Nickname: fmt.Sprintf("st_p_%v", id),
	}

	// 2. 序列化为 JSON
	jsonData, _ := json.Marshal(reqData)

	// 3. 发起 POST 请求
	url := "http://127.0.0.1:10081/api/player/create"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp, err := globalClient.Do(req)
	if err != nil {
		fmt.Println("请求失败:", err)
		atomic.AddInt32(&GFail, 1)
		return
	}
	defer resp.Body.Close() // 必须关闭响应体

	// 4. 读取响应
	ioutil.ReadAll(resp.Body)
	fmt.Println("状态码:", resp.StatusCode)
	//fmt.Println("响应内容:", string(body))
	atomic.AddInt32(&GSuccess, 1)
}

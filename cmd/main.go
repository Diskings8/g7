package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"g7/common/configx"
	"g7/common/etcd"
	"g7/common/globals"
	"g7/common/protocol"
	"g7/common/protos/pb"
	"g7/common/redisx"
	"github.com/go-redis/redis/v8"
	"os"
)

func main() {
	//fmt.Println(os.Getwd())
	globals.Env = globals.EnvTest
	configx.LoadEnvConf(globals.GetEnvConfPath())

	//checkRedis()
	checkEtcd()
}

func checkEtcd() {
	//test := string("123.207.11.230:32379")
	//etcd.InitETCD(test)
	etcd.InitETCD(configx.GEnvCfg.Etcd.Dsn)

	//checkEtcdGateway()
	//checkEtcdLogin()
	checkEtcdGame()
}

func checkEtcdGateway() {
	//etcd.UpdateEtcdConf(etcd_conf.ConfSwitchLoginOn, "true")
	for _, v := range etcd.GetServiceList(globals.GatewayRpc) {

		c, _ := protocol.NewGatewayNodeClient(context.Background(), v)
		rps, _ := c.GetConnCount(context.Background(), &pb.Req_Node_ConnCount{})
		fmt.Println(v, rps.GetCount())
	}
}

func checkEtcdGame() {
	//etcd.UpdateEtcdConf(etcd_conf.ConfSwitchLoginOn, "true")
	for _, v := range etcd.GetServiceList(globals.GameRpc) {

		fmt.Println(v)
	}
}

func checkEtcdLogin() {
	//etcd.UpdateEtcdConf(etcd_conf.ConfSwitchLoginOn, "true")
	for _, v := range etcd.GetServiceList(globals.LoginRpc) {

		fmt.Println(v)
	}
}

func checkRedis() {
	redisx.Init(configx.GEnvCfg.Redis.Addr, configx.GEnvCfg.Redis.Password, configx.GEnvCfg.Redis.DB)
	_ = redisx.SetKey("name", []byte("test key"), 0)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n输入 key：")
		scanner.Scan()
		key := scanner.Text()

		if key == "exit" {
			break
		}

		val, err := redisx.GetKey(key)
		if errors.Is(err, redis.Nil) {
			fmt.Println("key 不存在")
		} else if err != nil {
			fmt.Println("错误：", err)
		} else {
			fmt.Printf("✅ %s = %s\n", key, val)
		}
	}
}

package main

import (
	"context"
	"fmt"
	"g7/common/configx"
	"g7/common/etcd"
	"g7/common/globals"
	"g7/common/protocol"
	"g7/common/protos/pb"
)

func main() {
	//fmt.Println(os.Getwd())
	configx.LoadEnvConf(globals.ConfDev)
	etcd.InitETCD(configx.GEnvCfg.Etcd.Dsn)
	//etcd.UpdateEtcdConf(etcd_conf.ConfSwitchLoginOn, "true")
	for _, v := range etcd.GetServiceList(globals.GateWayServer) {

		c, _ := protocol.NewGatewayNodeClient(context.Background(), v)
		rps, _ := c.GetConnCount(context.Background(), &pb.Req_Node_ConnCount{})
		fmt.Println(v, rps.GetCount())
	}
}

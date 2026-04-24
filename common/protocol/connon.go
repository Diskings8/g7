package protocol

import (
	"context"
	"g7/common/protos/pb"
	"sync"
	"time"

	"google.golang.org/grpc"
)

var defaultPool = newNodeConnPool()

type NodeConn struct {
	Addr string
	*grpc.ClientConn
	// 这里可以存通用的健康检查状态
	Alive bool
}

type nodeConnPool struct {
	rwLock sync.RWMutex
	pool   map[string]*NodeConn // Key: 节点地址
}

func newNodeConnPool() *nodeConnPool {
	return &nodeConnPool{
		pool: make(map[string]*NodeConn),
	}
}

func (p *nodeConnPool) getOrCreateConn(addr string) (*NodeConn, error) {
	p.rwLock.RLock()
	nc, ok := p.pool[addr]
	p.rwLock.RUnlock()

	if ok {
		return nc, nil
	}

	// 双重检查锁
	p.rwLock.Lock()
	defer p.rwLock.Unlock()

	if nc, ok := p.pool[addr]; ok {
		return nc, nil
	}

	// 1. 建立底层 gRPC 连接
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	nc = &NodeConn{
		Addr:  addr,
		Alive: true,
	}
	nc.ClientConn = conn
	p.pool[addr] = nc
	return nc, nil
}

// 通用获取连接方法
func GetConn(addr string) (*NodeConn, error) {
	return defaultPool.getOrCreateConn(addr)
}

func NewGameNodeClient(addr string) (pb.GameNodeServiceClient, error) {
	// 1. 建立连接（单工不需要长流）
	conn, err := GetConn(addr)
	if err != nil {
		return nil, err
	}

	// 2. 创建单工客户端
	client := pb.NewGameNodeServiceClient(conn)
	return client, nil
}

func NewGameNodeStreamClient(addr string) (pb.GameStreamServiceClient, error) {
	// 1. 建立连接（单工不需要长流）
	conn, err := GetConn(addr)
	if err != nil {
		return nil, err
	}

	// 2. 创建单工客户端
	client := pb.NewGameStreamServiceClient(conn)
	return client, nil
}

func NewGatewayNodeClient(addr string) (pb.GatewayNodeServiceClient, error) {
	// 1. 建立连接（单工不需要长流）
	conn, err := GetConn(addr)
	if err != nil {
		return nil, err
	}

	// 2. 创建单工客户端
	client := pb.NewGatewayNodeServiceClient(conn)
	return client, nil
}

func NewMatchNodeClient(addr string) (pb.MatchNodeServiceClient, error) {
	// 1. 建立连接（单工不需要长流）
	conn, err := GetConn(addr)
	if err != nil {
		return nil, err
	}

	// 2. 创建单工客户端
	client := pb.NewMatchNodeServiceClient(conn)
	return client, nil
}

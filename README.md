# G7 游戏服务框架

本项目是一个基于 Go 的游戏服务端架构。包含登录服、网关服、游戏服、综合服、平台服、监控服务和辅助工具，支持 ETCD 服务发现、Redis 缓存、MySQL/MongoDB 存储、Kafka 消息以及 gRPC/RPC/TCP 通信。

## 目录结构

- `cmd/`：辅助工具入口，例如 ETCD/Redis 检查工具。
- `common/`：公共模块，包含配置、日志、ETCD、Redis、数据库、协议、JWT、限流、工具函数等。
- `game/`：游戏服实现，包含游戏逻辑、管理器、RPC 服务等。
- `gateway/`：网关服实现，负责 TCP 连接、转发、服务发现与注册。
- `login/`：登录服实现，包含认证、路由、MQ、数据库初始化等。
- `comprehensive/`：综合服入口，通常用于匹配、房间、管理等综合功能。
- `platform/`：平台服入口。
- `preposition/`：预置服务入口。
- `monitor/`：监控服务入口。
- `tcp_cli/`：TCP 客户端测试工具。
- `bin/`：编译/部署产物和容器环境配置。
- `config.yaml`：默认运行配置。
- `config_prod.yaml`：生产环境配置。

## 主要服务说明

### 登录服 (`login/main.go`)
- 使用 Gin 框架。
- 读取 `configx` 配置。
- 初始化 ETCD、Redis、Kafka、数据库、雪花 ID 以及路由。
- 在配置的登录端口上启动 HTTP 服务。

### 网关服 (`gateway/main.go`)
- 解析环境参数并加载配置。
- 初始化日志与 ETCD 注册。
- 启动 TCP 服务和 gRPC 服务。
- 负责客户端连接入口和后端游戏服路由。

### 游戏服 (`game/main.go`)
- 解析环境参数并加载配置。
- 初始化日志、雪花 ID、MySQL、Redis、MQ、ETCD、管理器和定时器。
- 注册 gRPC 服务并对外暴露游戏节点接口。

### 其他服务
- `comprehensive/`：综合模块入口，可能关联匹配、房间、系统管理。
- `platform/`：平台业务入口。
- `preposition/`：预置/准备类服务入口。
- `monitor/`：监控服务入口。
- `tcp_cli/`：TCP 客户端测试程序，用于快速连接服务进行测试。

## 依赖

- Go 1.24
- Gin
- GORM + MySQL
- Redis v8
- ETCD v3
- gRPC / Protocol Buffers
- Kafka (Sarama)
- MongoDB
- zap 日志
- JWT

## 配置说明

默认配置文件为 `config.yaml`。

常见配置项：
- `server`：平台、登录、游戏等服务端口。
- `mysql_global` / `mysql_game`：MySQL 连接。
- `mongodb_game`：MongoDB 连接。
- `redis`：Redis 地址与认证。
- `snowflake`：雪花 ID 参数。
- `jwt`：JWT 密钥和过期时间。
- `etcd`：ETCD 连接地址。
- `gateWay`：网关服务地址与 RPC 地址。
- `mq`：消息队列类型与地址。

环境配置由 `common/globals.GetEnvConfPath()` 自动解析。可通过 `-env`、`-platform`、`-container` 等参数传入运行时环境信息。

## 编译方式

### 一键编译

- `go_all_compile.bat`：依次构建游戏服、网关服、登录服，输出到 `bin/server/`。

### 单独编译

- `go_compile_game.bat`：构建 `game_server`。
- `go_compile_gateway.bat`：构建 `gateway_server`。
- `go_compile_login.bat`：构建 `login_server`。

这些脚本会设置：
- `CGO_ENABLED=0`
- `GOOS=linux`
- `GOARCH=amd64`

## 运行方式

### 本地测试

示例：

```bat
go run login/main.go -env=prod -platform=91 -container=local
go run gateway/main.go -env=prod -platform=91 -container=local
go run game/main.go -env=prod -server=1001 -platform=91 -container=local
```

### Docker / 容器模式

服务会读取环境变量 `POD_NAME`、`POD_IP` 等，用于构建 ETCD 注册地址。

### 参数说明

- `-env`：运行环境，例如 `test`、`prod`、`pre`。
- `-platform`：平台 ID，例如 `91`。
- `-container`：容器类型，例如 `local` 或 `docker`。
- `-server`：游戏服 ID（仅 `game` 服务使用）。

## 预备服务

启动项目前应确保以下基础服务可用：
- MySQL
- Redis
- ETCD
- Kafka（若使用 MQ）
- MongoDB（若使用 MongoDB 存储）

## 常见目录说明

- `common/configx/`：配置读取与环境配置管理。
- `common/etcd/`：服务发现与注册，支持网关、登录、游戏 RPC 路径注册。
- `common/redisx/`：Redis 连接封装。
- `common/dbc/`：数据库初始化与 ORM。
- `common/protocol/`：协议与 RPC 生成定义。
- `common/logger/`：日志初始化。
- `common/snowflakes/`：分布式 ID 生成。

## 备注

- 该项目使用的是 `go.mod` 管理依赖。
- `config_prod.yaml` 可作为生产环境配置参考。
- 如果后续要扩展服务，可以在 `common/` 中复用已有公共组件。

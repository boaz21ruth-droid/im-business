# im-business

`im-business` 是 Web1 IM 钱包业务后端，负责承接 Flutter 客户端的账号、用户资料、钱包、TOTP、P2P 与 Webhook 请求，并与 OpenIM Server、PostgreSQL、Redis 以及链上数据服务协同工作。

## 主要能力

- 账号体系：验证码、注册、登录、重置密码、修改密码。
- OpenIM 对接：通过 OpenIM 管理接口创建用户并签发 IM Token。
- 用户资料：查询、搜索、更新用户信息。
- 钱包业务：钱包地址登记、好友收款地址查询、交易历史、链上转账通知。
- Swap/Bridge/Intent：服务端聚合报价，客户端只拿可执行报价与运行时配置，不持有聚合器 API Key。
- TOTP：钱包敏感操作二次验证，密钥使用 AES-256 加密后落库。
- P2P：订单查询、买卖列表、争议处理接口，并可监听托管合约事件。
- Webhook：接收 Moralis、Alchemy、QuickNode 等流式交易通知。

## 技术栈

- Go 1.26.3
- Gin
- GORM + PostgreSQL
- Redis
- Zap
- Viper
- go-ethereum

## 目录结构

```text
cmd/server/              服务入口
config/                  示例配置
internal/config/         配置结构与环境变量覆盖
internal/db/             PostgreSQL / Redis 初始化
internal/handler/        HTTP 路由与请求处理
internal/model/          数据模型
internal/repo/           数据访问层
internal/service/        账号、OpenIM、TOTP 等业务服务
internal/wallet/         钱包、行情、交易历史、Swap、Bridge、Intent
internal/p2p/            P2P 合约监听
pkg/                     通用响应、JWT、验证码、加密、TOTP 工具
```

## 前置依赖

本服务默认依赖同仓库中的 OpenIM 基础设施：

- OpenIM Server API：默认 `http://localhost:10002`
- OpenIM WebSocket：由客户端直连，默认 `ws://host:10001`
- Redis：本地开发默认 `localhost:16379`
- PostgreSQL：业务库默认本地 `5432`，Docker 模式暴露为 `15432`

如使用 Docker Compose，需先启动 `open-im-server`，因为 `im-business/docker-compose.yaml` 会加入外部网络 `open-im-server_openim`。

## 本地配置

复制示例配置：

```bash
cd im-business
cp config/config.example.yaml config/config.yaml
```

至少需要检查这些配置：

- `postgres.dsn`：业务库连接串。
- `redis.addr` / `redis.password` / `redis.db`：建议使用 OpenIM 未占用的 DB，示例为 `5`。
- `jwt.secret`：生产环境必须替换为足够长的随机密钥。
- `totp.encrypt_key`：64 位十六进制 AES-256 Key，可用 `openssl rand -hex 32` 生成。
- `openim.api_url` / `openim.admin_user_id` / `openim.secret`：需与 OpenIM Server 配置一致。
- `wallet.*`：链上交易历史、Webhook、RPC 与 Provider API Key。
- `swap.*`：0x 等报价服务配置、每条链 RPC、路由白名单、平台费接收地址和风控阈值。

配置支持环境变量覆盖，字段中的点会映射成下划线。例如：

```bash
POSTGRES_DSN="host=localhost user=im_business password=xxx dbname=im_business port=5432 sslmode=disable" \
SERVER_MODE=release \
go run ./cmd/server -config config/config.yaml
```

## 启动

本地直接运行：

```bash
cd im-business
go mod download
go run ./cmd/server -config config/config.yaml
```

健康检查：

```bash
curl http://localhost:10008/healthz
```

Docker Compose 运行：

```bash
cd ../open-im-server
docker compose up -d

cd ../im-business
cp config/config.example.yaml config/config.yaml
docker compose up -d --build
```

Docker 模式下，Compose 会覆盖容器内访问地址：

- PostgreSQL：`postgres-business:5432`
- Redis：`redis:6379`
- OpenIM API：`http://openim-server:10002`
- im-business：宿主机 `http://localhost:10008`

## 测试

```bash
cd im-business
go test ./...
```

## 接口分组

公开接口：

- `GET /healthz`
- `GET /app/check`
- `POST /client_config/get`
- `POST /account/code/send`
- `POST /account/code/verify`
- `POST /account/register`
- `POST /account/login`
- `POST /account/password/reset`
- `POST /webhook/:provider`
- `GET /p2p/orders`
- `GET /p2p/orders/:id`

需要登录态的接口使用 `token` 请求头，或 `Authorization: Bearer <token>`：

- `POST /account/password/change`
- `POST /user/update`
- `POST /user/find/full`
- `POST /user/search/full`
- `POST /user/rtc/get_token`
- `POST /user/totp/setup`
- `POST /user/totp/enable`
- `POST /user/totp/disable`
- `GET /user/totp/status`
- `POST /friend/search`
- `POST /wallet/addresses`
- `GET /wallet/addresses`
- `GET /wallet/friend/address`
- `GET /wallet/tx-history`
- `GET /wallet/swap_config`
- `GET /wallet/price`
- `GET /wallet/quote`
- `GET /wallet/bridge/quote`
- `GET /wallet/bridge/status`
- `GET /wallet/intent/quote`
- `POST /wallet/intent/order`
- `GET /wallet/intent/status`
- `POST /wallet/totp/verify`
- `GET /p2p/my/selling`
- `GET /p2p/my/buying`
- `GET /p2p/admin/disputes`
- `POST /p2p/admin/resolve`

## 与 im-wallet-app 联调

客户端默认通过 `openim_common/lib/src/config.dart` 访问：

- 业务接口：`/chat` 或直连 `:10008`
- OpenIM API：`/api` 或直连 `:10002`
- OpenIM WebSocket：`/msg_gateway` 或直连 `:10001`

如果没有反向代理，客户端需切到直连端口模式，或在 App 运行时写入 server config。开发机、手机和模拟器必须能访问同一个局域网地址。

## 生产注意事项

- 不要提交真实的 `config/config.yaml`、JWT Secret、TOTP Key、Provider API Key 或私钥。
- `p2p` 管理接口目前只做登录认证，生产环境应增加管理员权限和 IP 白名单。
- Webhook URL 需要是外部可访问的 HTTPS 地址，并配置服务商签名密钥。
- `server.mode` 生产环境使用 `release`。
- 建议在反向代理层限制 CORS、请求体大小、频率和管理接口来源。

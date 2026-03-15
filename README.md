# Wallet Service

基于 Go 语言开发的轻量级钱包服务，采用 Clean Architecture 架构设计，支持多种存储后端。

## 功能特性

- 创建钱包
- 查询钱包余额
- 钱包间转账（支持并发操作）
- 可切换存储模式（内存/MySQL）
- 支持配置文件和环境变量

## 技术栈

- Go 1.25+
- Gin Web 框架
- Viper 配置管理
- GORM MySQL 驱动
- 内存存储（分段锁实现并发安全）

## 项目结构

```
.
├── cmd/
│   └── server/           # 服务入口
├── config/               # 配置文件
├── internal/
│   ├── controller/       # HTTP 控制器
│   ├── domain/          # 领域模型和错误定义
│   ├── id/              # ID 生成器
│   ├── repo/            # 数据仓储层
│   │   ├── memoryRepo.go   # 内存存储实现
│   │   └── mysqlRepo.go    # MySQL 存储实现
│   ├── router/          # 路由配置
│   └── service/         # 业务逻辑层
```

## 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd Wallet
```

### 2. 运行服务

```bash
go run ./cmd/server/main.go
```

服务启动后默认监听 `http://localhost:8080`

### 3. API 接口

| 方法 | 路径 | 说明 | 请求体 |
|------|------|------|--------|
| POST | /wallets | 创建钱包 | - |
| GET | /wallets/:id | 获取钱包信息 | - |
| POST | /transfer | 转账 | `{"source_id": "w_1", "destination_id": "w_2", "amount": 100}` |

### 4. 示例

```bash
# 创建钱包
curl -X POST http://localhost:8080/wallets

# 查询钱包
curl http://localhost:8080/wallets/w_1

# 转账
curl -X POST http://localhost:8080/transfer \
  -H "Content-Type: application/json" \
  -d '{"source_id": "w_1", "destination_id": "w_2", "amount": 100}'
```

## 配置说明

配置文件位于 `config/config.yaml`，支持以下配置项：

```yaml
server:
  port: 8080              # 服务端口

database:
  host: localhost         # MySQL 主机
  port: 3306              # MySQL 端口
  user: root             # 用户名
  password: ""           # 密码
  dbname: wallet         # 数据库名

repository:
  type: memory            # 存储类型: memory 或 mysql
  segment_count: 64       # 分段数量（仅 memory 模式有效）
```

### 环境变量

支持通过环境变量覆盖配置：

| 环境变量 | 说明 |
|----------|------|
| WALLET_SERVER_PORT | 服务端口 |
| WALLET_REPOSITORY_TYPE | 存储类型 |
| WALLET_REPOSITORY_SEGMENT_COUNT | 分段数量 |
| WALLET_DATABASE_HOST | MySQL 主机 |
| WALLET_DATABASE_PORT | MySQL 端口 |
| WALLET_DATABASE_USER | MySQL 用户 |
| WALLET_DATABASE_PASSWORD | MySQL 密码 |
| WALLET_DATABASE_DBNAME | 数据库名 |

### MySQL 模式

切换到 MySQL 存储：

```yaml
repository:
  type: mysql
```

确保 MySQL 数据库已创建：

```sql
CREATE DATABASE wallet;
```

服务启动时会自动创建 `wallets` 表。

## 测试

```bash
# 运行所有测试
go test ./...

# 运行单个测试
go test ./internal/service/... -v -run TestWalletService_Transfer

# 带竞态检测的测试
go test -race ./...
```

## 架构设计

### 分段锁并发控制（Memory 模式）

MemoryRepo 使用分段锁（Shard Locking）策略：
- 默认 64 个分段，每个分段独立的 RWMutex
- 钱包 ID 通过 FNV 哈希映射到对应分段
- 同一分段内转账只需锁定单个分段
- 跨分段转账按索引顺序加锁避免死锁

这种方式在高并发场景下能有效减少锁竞争。

### 事务行锁（MySQL 模式）

MySQLRepo 使用 `SELECT FOR UPDATE` 实现分布式锁：
- 事务内按 ID 顺序锁定钱包记录
- 使用原子更新 `balance = balance - amount` 防止超额转账
- 数据库保证事务的 ACID 特性
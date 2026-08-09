# Slink

自托管个人图床:Go(Gin)+ Vue 3,单二进制部署,支持多存储后端、多数据库与多级缓存。

## 功能特性

- **图片上传**:拖拽/粘贴/多文件/URL 拉取,失败可重试;页内最大化预览
- **MD5 去重**:相同内容只落盘一次,数据库记录共享同一文件与 URL
- **原样存储**:不转码不压缩,落盘文件与原图格式、MD5 完全一致
- **多存储后端**:本地、S3、MinIO、OSS、COS、七牛、又拍云、WebDAV、FTP、SFTP
- **多数据库**:SQLite(零依赖默认)/ MySQL / PostgreSQL / SQL Server
- **多级缓存**:内存 / 文件 / Redis;用户与配置读取走 L1(进程内)+ L2(Redis) 两级缓存
- **多用户与权限**:注册开关、管理员后台、按用户分配存储策略与上传策略组(限流/命名规则/大小限制)、存储容量配额
- **图片分享**:密码、有效期、最大查看次数
- **备份恢复**:全量导出/导入(config + 数据库 + 本地图片)

## 快速开始

### 本地运行(默认 SQLite,零依赖)

```bash
cd SlinkWeb && pnpm install && pnpm build && cd ..
go build -o slink .
./slink          # 监听 :8080,首次打开进入初始化向导
```

### Docker(SQLite 单容器)

```bash
cd SlinkWeb && pnpm install && pnpm build && cd ..
SLINK_JWT_SECRET=$(openssl rand -base64 48) docker compose up -d --build
```

### Docker Compose(PostgreSQL + Redis 全量)

```bash
cd SlinkWeb && pnpm install && pnpm build && cd ..
# 准备 .env:SLINK_JWT_SECRET / POSTGRES_PASSWORD 等(见 docker-compose.postgres.yml)
docker compose -f docker-compose.postgres.yml --env-file .env up -d --build
```

初始化向导中数据库选 `postgres`(主机 `postgres`,端口 5432),缓存选 `redis`(主机 `redis`,端口 6379)。

### 环境变量

| 变量 | 说明 | 默认 |
|---|---|---|
| `SLINK_ADDR` | 监听地址 | `:8080` |
| `SLINK_JWT_SECRET` | JWT 签名密钥(**生产必设长随机串**) | 内置开发值 |
| `DB_TYPE` / `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | 数据库(无 `config/database.json` 时生效) | sqlite / data.db |
| `CACHE_TYPE` / `REDIS_HOST` / `REDIS_PORT` / `REDIS_PASSWORD` / `REDIS_DB` | 缓存(无 `config/cache.json` 时生效) | memory |

## 性能测试

### 实测环境

Apple Silicon(M 系列)/ 16GB,应用裸机运行;PostgreSQL 18 与 Redis 8 运行于 Docker Desktop(存在虚拟化网络往返,每次 DB 查询约 +1ms——**这是上传链路的主要限制,Linux 同宿部署会显著更快**,见下方估算)。

| 场景 | 并发 | QPS | p50 | p95 | p99 | 应用内存峰值 |
|---|---|---|---|---|---|---|
| 图片访问 | 50 | 12,721 | 1.9ms | 11.2ms | 75.0ms | 41MB |
| 图片访问 | 200 | **18,302** | 2.9ms | 15.5ms | 316ms | 48MB |
| 上传(64KB) | 10 | **519** | 15.6ms | 34.3ms | 44.0ms | 48MB |
| 上传(1MB) | 10 | 168 | 54.0ms | 111ms | 119ms | 94MB |
| 图片列表 API | 20 | 1,018 | 11.7ms | 55.2ms | 79.2ms | 95MB |

资源占用:应用常驻 ~40MB;PostgreSQL ~75MB;Redis ~36MB。

### Linux 服务器估算

以下为本机实测结合"消除 macOS Docker 虚拟化网络往返(每次 DB/Redis 查询省约 0.8~1ms)"推算的**估算值,非实测**,仅供容量参考;实际受 CPU 单核性能、磁盘 IOPS、虚拟化方式影响:

| 规格 | 图片访问 QPS | 上传 QPS(64KB) | 图片列表 QPS | 说明 |
|---|---|---|---|---|
| 2C4G VPS | 8,000 ~ 12,000 | 700 ~ 1,000 | 800 ~ 1,200 | 可流畅跑全栈(App+PG+Redis) |
| 4C8G | 15,000 ~ 20,000 | 1,200 ~ 2,000 | 1,500 ~ 2,500 | 推荐生产规格 |
| 8C16G | 25,000 ~ 35,000 | 2,000 ~ 3,500 | 2,500 ~ 4,000 | 上传瓶颈转向 PG 提交 fsync |

内存基线估算:App ~60MB + PostgreSQL ~200MB + Redis ~50MB,**2GB 内存即可起步**,4GB 宽裕。

### 复现压测

```bash
# 上传(每个请求内容唯一,避免去重干扰;-token 为登录 JWT)
go run ./tools/loadtest -mode upload -url http://127.0.0.1:8080/api/image/upload -token <JWT> -n 300 -c 10 -size 64

# 图片访问
go run ./tools/loadtest -mode get -url http://127.0.0.1:8080/static/<图片路径> -n 10000 -c 200

# 图片列表(鉴权 + DB)
go run ./tools/loadtest -mode get -url "http://127.0.0.1:8080/api/image/list?page=1&limit=20" -token <JWT> -n 1000 -c 20
```

输出 QPS 与 p50/p95/p99/max 延迟;服务端 CPU/内存可在压测期间用 `ps -o %cpu=,rss= -p <PID>` 采样。

## 目录结构

```
├── main.go / router.go    # 入口与路由
├── api/                   # HTTP 接口(上传/用户/管理/分享/备份)
├── middleware/            # JWT/Bearer 鉴权、管理员校验、防盗链、用户缓存
├── model/                 # GORM 模型与各数据库初始化
├── storage/               # 存储后端(local/s3/oss/cos/qiniu/upyun/webdav/ftp/sftp/minio)
├── cache/                 # 内存/文件/Redis 缓存
├── utils/                 # 文件保存、链接生成等
├── tools/loadtest/        # 压测工具
└── SlinkWeb/              # Vue 3 + Vite 前端(构建产物 dist 由后端直接托管)
```

## 安全说明

- 生产部署务必设置 `SLINK_JWT_SECRET` 长随机串;
- 管理员账号由初始化向导创建,系统不会生成任何默认/后门账号;
- 管理接口由 `IsAdmin` 强制校验,普通注册用户无法访问。

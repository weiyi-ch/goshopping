# Go-Zero-Mall开发日记

原项目地址：https://github.com/nivin-studio/go-zero-mall

文档地址：https://juejin.cn/post/7036011047391592485

目标：增加原文档shop处理高并发功能

## P1-环境搭建

### 0.docker工作流程

镜像好比"模具"，容器好比"用模具生产的产品"，docker-compose.yml 好比"生产线说明书"，指导镜像变成容器并连接。

1.dockerfile负责构建镜像

```
Dockerfile → 构建 → 镜像（Image）
    ↓
多个镜像 → 组成 → 镜像仓库
```

2.docker-compose.yml负责通过这些镜像生成容器并构建多容器之间的通信。

```
docker-compose.yml → 编排 → 容器组
    ↓
1. 创建共享网络
2. 按顺序启动容器
3. 容器间自动连接
4. 端口映射到主机
```

### 1.创建文件目录

```bash
# 创建主目录
mkdir -p goweiyishop
# 进入主目录
cd goweiyishop
# 创建各个子目录
mkdir -p dtm
mkdir -p etcd
mkdir -p etcd-manage
mkdir -p golang
mkdir -p grafana
mkdir -p jaeger
mkdir -p mysql
mkdir -p mysql-manage
mkdir -p prometheus
mkdir -p redis
mkdir -p redis-manage
# 创建配置文件
touch .env
touch docker-compose.yml
# 在各个子目录中创建对应文件
touch dtm/config.yml
touch dtm/Dockerfile
touch etcd/Dockerfile
touch etcd-manage/Dockerfile
touch golang/Dockerfile
touch grafana/Dockerfile
touch jaeger/Dockerfile
touch mysql/Dockerfile
touch mysql-manage/Dockerfile
touch prometheus/Dockerfile
touch prometheus/prometheus.yml
touch redis/Dockerfile
touch redis-manage/Dockerfile
# 验证目录结构
echo "目录结构创建完成，当前结构："
find . -type f | sort
```

### 2.配置基础Go环境Dockerfile

goctl 最新版本要求 Go >= 1.23，grpc要求1.24，原文档中1.18不符合要求，

```dockerfile
FROM golang:1.25

LABEL maintainer="Weiyi <15671108020@163.com>"

ENV GOPROXY https://goproxy.cn,direct

# 安装必要的软件包和依赖包
USER root
RUN sed -i 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/' /etc/apt/sources.list && \
    sed -i 's/security.debian.org/mirrors.tuna.tsinghua.edu.cn/' /etc/apt/sources.list && \
    sed -i 's/security-cdn.debian.org/mirrors.tuna.tsinghua.edu.cn/' /etc/apt/sources.list && \
    apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends \
    curl \
    zip \
    unzip \
    git \
    vim 

# 安装 goctl
USER root
RUN GOPROXY=https://goproxy.cn/,direct go install github.com/zeromicro/go-zero/tools/goctl@latest

# 安装 protoc
USER root
RUN curl -L -o /tmp/protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.19.1/protoc-3.19.1-linux-x86_64.zip && \
    unzip -d /tmp/protoc /tmp/protoc.zip && \
    mv /tmp/protoc/bin/protoc $GOPATH/bin

# 安装 protoc-gen-go
USER root
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# 安装 protoc-gen-go-grpc
USER root
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# $GOPATH/bin添加到环境变量中
ENV PATH $GOPATH/bin:$PATH

# 清理垃圾
USER root
RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* && \
    rm /var/log/lastlog /var/log/faillog

# 设置工作目录
WORKDIR /usr/src/code

EXPOSE 8000
EXPOSE 8001
EXPOSE 8002
EXPOSE 8003
EXPOSE 9000
EXPOSE 9001
EXPOSE 9002
EXPOSE 9003

```

### 3.修改dockerfile

```dockerfile
#大部分直接用已有镜像
FROM --platform=linux/arm64 yedf/dtm:latest
```

### 4.设置环境变量

```dockerfile
#在编排容器中定义环境变量
# 设置时区
TZ=Asia/Shanghai
# 设置网络模式
NETWORKS_DRIVER=bridge


# PATHS ##########################################
# 宿主机上代码存放的目录路径
CODE_PATH_HOST=./code
# 宿主机上Mysql Reids数据存放的目录路径
DATA_PATH_HOST=./data


# MYSQL ##########################################
# Mysql 服务映射宿主机端口号，可在宿主机127.0.0.1:3306访问
MYSQL_PORT=3306
MYSQL_USERNAME=admin
MYSQL_PASSWORD=123456
MYSQL_ROOT_PASSWORD=123456

# Mysql 可视化管理用户名称，同 MYSQL_USERNAME
MYSQL_MANAGE_USERNAME=admin
# Mysql 可视化管理用户密码，同 MYSQL_PASSWORD
MYSQL_MANAGE_PASSWORD=123456
# Mysql 可视化管理ROOT用户密码，同 MYSQL_ROOT_PASSWORD
MYSQL_MANAGE_ROOT_PASSWORD=123456
# Mysql 服务地址
MYSQL_MANAGE_CONNECT_HOST=mysql
# Mysql 服务端口号
MYSQL_MANAGE_CONNECT_PORT=3306
# Mysql 可视化管理映射宿主机端口号，可在宿主机127.0.0.1:1000访问
MYSQL_MANAGE_PORT=1000


# REDIS ##########################################
# Redis 服务映射宿主机端口号，可在宿主机127.0.0.1:6379访问
REDIS_PORT=6379

# Redis 可视化管理用户名称
REDIS_MANAGE_USERNAME=admin
# Redis 可视化管理用户密码
REDIS_MANAGE_PASSWORD=123456
# Redis 服务地址
REDIS_MANAGE_CONNECT_HOST=redis
# Redis 服务端口号
REDIS_MANAGE_CONNECT_PORT=6379
# Redis 可视化管理映射宿主机端口号，可在宿主机127.0.0.1:2000访问
REDIS_MANAGE_PORT=2000


# ETCD ###########################################
# Etcd 服务映射宿主机端口号，可在宿主机127.0.0.1:2379访问
ETCD_PORT=2379
# 不要用7000，mac系统7000被占用无法正常访问
# ETCD_MANAGE_PORT=7000
ETCD_MANAGE_PORT=7001
# PROMETHEUS #####################################
# Prometheus 服务映射宿主机端口号，可在宿主机127.0.0.1:3000访问
PROMETHEUS_PORT=3000


# GRAFANA ########################################
# Grafana 服务映射宿主机端口号，可在宿主机127.0.0.1:4000访问
GRAFANA_PORT=4000


# JAEGER #########################################
# Jaeger 服务映射宿主机端口号，可在宿主机127.0.0.1:5000访问
JAEGER_PORT=5000


# DTM #########################################
# DTM HTTP 协议端口号
DTM_HTTP_PORT=36789
# DTM gRPC 协议端口号
DTM_GRPC_PORT=36790

```

### 5.编排容器

运行时可能存在环境或者镜像版本号问题自查

```yaml
version: '3.5'
# 网络配置
networks:
   backend:
      driver: ${NETWORKS_DRIVER}

# 服务容器配置
services:
   golang: # 自定义容器名称
      build:
         context: golang                  # 指定构建使用的 Dockerfile 文件
      environment: # 设置环境变量
         - TZ=${TZ}
      privileged: true
      volumes: # 设置挂载目录
         - ${CODE_PATH_HOST}:/usr/src/code  # 引用 .env 配置中 CODE_PATH_HOST 变量，将宿主机上代码存放的目录挂载到容器中 /usr/src/code 目录
      ports: # 设置端口映射
         - "8000:8000"
         - "8001:8001"
         - "8002:8002"
         - "8003:8003"
         - "9000:9000"
         - "9001:9001"
         - "9002:9002"
         - "9003:9003"
      stdin_open: true                     # 打开标准输入，可以接受外部输入
      tty: true
      networks:
         - backend
      restart: always                      # 指定容器退出后的重启策略为始终重启

   etcd: # 自定义容器名称
      build:
         context: etcd                    # 指定构建使用的 Dockerfile 文件
      environment:
         - TZ=${TZ}
         - ALLOW_NONE_AUTHENTICATION=yes
         - ETCD_ADVERTISE_CLIENT_URLS=http://etcd:2379
      ports: # 设置端口映射
         - "${ETCD_PORT}:2379"
      networks:
         - backend
      restart: always

   mysql:
      build:
         context: mysql
      environment:
         - TZ=${TZ}
         - MYSQL_USER=${MYSQL_USERNAME}                  # 设置 Mysql 用户名称
         - MYSQL_PASSWORD=${MYSQL_PASSWORD}              # 设置 Mysql 用户密码
         - MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}    # 设置 Mysql root 用户密码
      privileged: true
      volumes:
         - ${DATA_PATH_HOST}/mysql:/var/lib/mysql        # 引用 .env 配置中 DATA_PATH_HOST 变量，将宿主机上存放 Mysql 数据的目录挂载到容器中 /var/lib/mysql 目录
      ports:
         - "${MYSQL_PORT}:3306"                          # 设置容器3306端口映射指定宿主机端口
      networks:
         - backend
      restart: always

   redis:
      build:
         context: redis
      environment:
         - TZ=${TZ}
      privileged: true
      volumes:
         - ${DATA_PATH_HOST}/redis:/data                 # 引用 .env 配置中 DATA_PATH_HOST 变量，将宿主机上存放 Redis 数据的目录挂载到容器中 /data 目录
      ports:
         - "${REDIS_PORT}:6379"                          # 设置容器6379端口映射指定宿主机端口
      networks:
         - backend
      restart: always

   mysql-manage:
      build:
         context: mysql-manage
      environment:
         - TZ=${TZ}
         - PMA_ARBITRARY=1
         - MYSQL_USER=${MYSQL_MANAGE_USERNAME}               # 设置连接的 Mysql 服务用户名称
         - MYSQL_PASSWORD=${MYSQL_MANAGE_PASSWORD}           # 设置连接的 Mysql 服务用户密码
         - MYSQL_ROOT_PASSWORD=${MYSQL_MANAGE_ROOT_PASSWORD} # 设置连接的 Mysql 服务 root 用户密码
         - PMA_HOST=${MYSQL_MANAGE_CONNECT_HOST}             # 设置连接的 Mysql 服务 host，可以是 Mysql 服务容器的名称，也可以是 Mysql 服务容器的 ip 地址
         - PMA_PORT=${MYSQL_MANAGE_CONNECT_PORT}             # 设置连接的 Mysql 服务端口号
      ports:
         - "${MYSQL_MANAGE_PORT}:80"                         # 设置容器80端口映射指定宿主机端口，用于宿主机访问可视化web
      depends_on: # 依赖容器
         - mysql                                             # 在 Mysql 服务容器启动后启动
      networks:
         - backend
      restart: always

   redis-manage:
      build:
         context: redis-manage
      environment:
         - TZ=${TZ}
         - ADMIN_USER=${REDIS_MANAGE_USERNAME}           # 设置 Redis 可视化管理的用户名称
         - ADMIN_PASS=${REDIS_MANAGE_PASSWORD}           # 设置 Redis 可视化管理的用户密码
         - REDIS_1_HOST=${REDIS_MANAGE_CONNECT_HOST}     # 设置连接的 Redis 服务 host，可以是 Redis 服务容器的名称，也可以是 Redis 服务容器的 ip 地址
         - REDIS_1_PORT=${REDIS_MANAGE_CONNECT_PORT}     # 设置连接的 Redis 服务端口号
      ports:
         - "${REDIS_MANAGE_PORT}:80"                     # 设置容器80端口映射指定宿主机端口，用于宿主机访问可视化web
      depends_on: # 依赖容器
         - redis                                         # 在 Redis 服务容器启动后启动
      networks:
         - backend
      restart: always

   etcd-manage:
      build:
         context: etcd-manage
      environment:
         - TZ=${TZ}
      ports:
         - "${ETCD_MANAGE_PORT}:8080"                    # 设置容器8080端口映射指定宿主机端口，用于宿主机访问可视化web
      depends_on: # 依赖容器
         - etcd                                          # 在 etcd 服务容器启动后启动
      networks:
         - backend
      restart: always

   prometheus:
      build:
         context: prometheus
      environment:
         - TZ=${TZ}
      privileged: true
      volumes:
         - ./prometheus/prometheus.yml:/opt/bitnami/prometheus/conf/prometheus.yml  # 将 prometheus 配置文件挂载到容器里
      ports:
         - "${PROMETHEUS_PORT}:9090"                     # 设置容器9090端口映射指定宿主机端口，用于宿主机访问可视化web
      networks:
         - backend
      restart: always

   grafana:
      build:
         context: grafana
      environment:
         - TZ=${TZ}
      ports:
         - "${GRAFANA_PORT}:3000"                        # 设置容器3000端口映射指定宿主机端口，用于宿主机访问可视化web
      networks:
         - backend
      restart: always

   jaeger:
      build:
         context: jaeger
      environment:
         - TZ=${TZ}
      ports:
         - "${JAEGER_PORT}:16686"                        # 设置容器16686端口映射指定宿主机端口，用于宿主机访问可视化web
      networks:
         - backend
      restart: always

   dtm:
      build:
         context: dtm
      environment:
         - TZ=${TZ}
      entrypoint:
         - "/app/dtm/dtm"
         - "-c=/app/dtm/configs/config.yaml"
      privileged: true
      volumes:
         - ./dtm/config.yml:/app/dtm/configs/config.yaml # 将 dtm 配置文件挂载到容器里
      ports:
         - "${DTM_HTTP_PORT}:36789"
         - "${DTM_GRPC_PORT}:36790"
      networks:
         - backend
      restart: always

```

## P2-服务划分

### 1.服务划分

**“服务拆分”**：将一个商城项目拆分为多个独立的微服务（User, Product, Order, Pay），每个服务都包含 API 层和 RPC 层，以实现高内聚、低耦合的架构。

用户服务（User）

产品服务（Product）

订单服务（Order）

支付服务（Pay）

### 2. **服务分层设计（每个服务）**

每个微服务都包含两个子服务：

**API 服务**：对外暴露接口，供前端 App 或其他外部系统调用。

**RPC 服务**：对内提供服务，供内部其他 API 服务或 RPC 服务调用。

### 3. **服务端口分配**

每个服务都有独立的端口规划，避免冲突：


| 服务     | API 服务端口 | RPC 服务端口 |
| -------- | ------------ | ------------ |
| 用户服务 | 8000         | 9000         |
| 产品服务 | 8001         | 9001         |
| 订单服务 | 8002         | 9002         |
| 支付服务 | 8003         | 9003         |

### 4. **接口示例**

每个服务都规划了核心接口：

**用户服务**：

login（登录）

register（注册）

userinfo（用户信息）

**产品服务**：

create（创建产品）

update（修改产品）

remove（删除产品）

detail（产品详情）

**订单服务**：

create（创建订单）

update（修改订单）

remove（删除订单）

detail（订单详情）

list（订单列表）

paid（支付接口）

**支付服务**：

create（创建支付）

detail（支付详情）

callback（支付回调）

---

### 5.**项目目录结构设计**

作者提供了标准的项目目录结构，便于团队开发和扩展：

```
mall/
├── common/           # 通用库（如工具函数、中间件、配置等）
├── service/          # 所有微服务目录
│   ├── user/
│   │   ├── api/      # 用户 API 服务代码
│   │   ├── model/    # 用户数据模型
│   │   └── rpc/      # 用户 RPC 服务代码
│   ├── product/
│   │   ├── api/
│   │   ├── model/
│   │   └── rpc/
│   ├── order/
│   │   ├── api/
│   │   ├── model/
│   │   └── rpc/
│   └── pay/
│       ├── api/
│       ├── model/
│       └── rpc/
└── go.mod            # Go 模块管理文件
```

## P3-用户服务

### 1.go-zero 自动生成代码的架构深度设计

```
1.纵向三层分层模型：
```

```

```

**API 层 (Gateway)**：对外暴露的 HTTP 接口，负责路由转发、参数校验及响应封装。

```

```

**RPC 层 (Service)**：核心业务逻辑所在地，负责数据持久化、跨服务调用及事务控制。

```

```

**Model 层 (Data Access)**：数据库操作的抽象，包含自动集成的 Redis 一致性缓存。

```
2.
```

`ServiceContext`：微服务的资源中心

```
代码启动时生成一个svcCtx，避免数据库与 Redis 连接数在并发下爆炸。
```

```
原理：
```

```markdown
这背后主要依靠两个机制：生命周期管理 和 内置连接池。

	A. 对象的单例化（Singleton）

	由于 `ServiceContext` 只在服务启动时被创建一次，它持有的 `sqlx.SqlConn` 或 `redis.Redis` 实际上是单例。

- **请求 A** 进来：使用 `svcCtx.UserModel`。
- **请求 B** 进来：也使用 **同一个** `svcCtx.UserModel`。 它们共同指向底层同一个数据库驱动实例。

	B. 连接池（Connection Pooling）

	这是最核心的技术细节。`go-zero` 底层使用的 `sqlx` 包装了 Go 原生的 `database/sql`。

- **什么是连接池**：当你初始化 `UserModel` 时，底层并没有只建立“一个”连接，而是建立了一个 **“池子”**。
- **工作流程**：
  1. 当一个业务逻辑需要执行 SQL 时，它去池子里 **“借用”** 一个空闲的 TCP 连接。
  2. SQL 执行完后，连接 **不会被断开**，而是 **“归还”** 到池子里。
  3. 下一个请求进来，直接复用池子里的这个连接。
```

### 2.API 与 RPC 的通信机制与链路细节

#### 2.1 服务发现与注册

在分布式系统中，API 不知道 RPC 的具体 IP 地址，它们通过 Etcd 进行“挂号”与“查号”。

- **RPC 注册**：`s.Start()` 启动时，`zrpc` 框架会自动向 Etcd 申请租约，并将服务名与当前容器 IP 绑定（例如：`user.rpc -> 192.168.1.5:9000`）。
- **API 发现**：API 启动时根据配置的 Etcd 地址，实时监听 `user.rpc` 下的节点变动。
- **健康检查**：Etcd 维持心跳，一旦 RPC 容器宕机，Etcd 会在秒级内通知 API 移除该节点，实现**自动故障转移**。

#### 2.2 客户端封装：`userclient` 的解耦艺术

API 调用 RPC 时，并不是直接写 gRPC 原始代码，而是通过 `userclient`：

- **原理**：`goctl` 生成的 `userclient` 将底层的连接管理、Proto 协议转换、负载均衡算法（P2C）全部封装。
- **代码表现**：在 API 的 Logic 层，只需一行 `l.svcCtx.UserRpc.Login(...)`，就像调用本地函数一样简单。

---

### 3. 从描述文件到完整代码

当你手握 `.api` 和 `.proto` 文件时，完成一个功能的标准化路径如下：

#### 3.1 代码生成（工具先行）

使用 `goctl` 命令行工具生成脚手架：

1. **生成 RPC**：`goctl rpc protoc user.proto --go_out=./pb --go-grpc_out=./pb --zrpc_out=.`
2. **生成 API**：`goctl api go -api user.api -dir .`
3. **生成 Model**：`goctl model mysql datasource -url "root:pass@tcp(127.0.0.1:3306)/db" -table "user" -dir ./model -c` ( `-c` 代表开启缓存)

#### 3.2 依赖注入（资源归仓）

在生成的 `internal/svc/servicecontext.go` 中，手动补齐你需要共享的资源：

- 将 `UserModel` 注入到 `ServiceContext`。
- 在 API 的 `ServiceContext` 中初始化 `UserRpc` 客户端。

#### 3.3 逻辑编写（业务填充）

在 `internal/logic` 目录下找到对应的文件：

1. **RPC Logic**：编写真正的数据库操作（调用 `m.UserModel.FindOne` 等）。
2. **API Logic**：调用 RPC 接口，并对返回数据进行 HTTP 层的业务封装（如 JWT 生成）。

---

### 4. 环境调试：

#### 4.1 网络监听

- **教训**：容器化部署时，服务（Etcd/MySQL/Manage）必须监听 `0.0.0.0`。
- **原因**：监听 `127.0.0.1` 意味着只接受来自容器内部的请求，会导致 Docker 端口映射失效。

#### 4.2 端口冲突

- **发现**：Mac 系统的 **AirPlay 接收器** 默认占用 7000 端口。
- **对策**：开发环境建议避开系统保留端口，使用 `7001` 或 `8001` 等，或在系统设置中关闭 AirPlay 接收器。

#### 4.3 数据库权限与持久化

- **权限**：Docker 初始化的普通用户通常没有 `Create Database` 权限，初始化脚本或管理操作建议使用 `root`。
- **持久化**：MySQL 和 Redis 的数据必须挂载到宿主机（`volumes`），否则容器重启后数据会全丢。

## P4-商品服务

### 1.完善基础代码

追踪查询逻辑

#### 1.`FindOne` 全链路执行路径

当你调用 `res, err := l.svcCtx.ProductModel.FindOne(...)` 时，程序走过的“套路”如下：


| **阶段**          | **所在组件/文件**               | **核心职责**                                        |
| ----------------- | ------------------------------- | --------------------------------------------------- |
| **1. 业务逻辑层** | `logic/detail_logic.go`         | 发起业务查询请求。                                  |
| **2. 模型接口层** | `model/product_model_gen.go`    | 生成 Redis Key，定义回马枪（查库匿名函数）。        |
| **3. 缓存编排层** | `sqlc/cachedsql.go`             | 统一入口，决定是否启用缓存逻辑。                    |
| **4. 集群调度层** | `cache/cluster.go`              | 通过**一致性哈希**决定请求去哪台 Redis 节点。       |
| **5. 策略执行层** | `cache/cache.go` (**`doTake`**) | **核心大脑**：处理 SingleFlight、防穿透、查库回填。 |
| **6. 物理执行层** | `sqlx/sqlconn.go`               | 最终执行 SQL，处理结果集映射（Unmarshal）。         |

---

#### 2.关键源码：`doTake` 深度注释版

**文件路径**：`core/stores/cache/cache.go`

Go

```golang
// doTake 是 go-zero 缓存设计的定海神针
func (c cacheNode) doTake(ctx context.Context, v interface{}, key string,
    query func(v interface{}) error, cacheVal func(v interface{}) error) error {
  
    // 【1. 屏障拦截 - SingleFlight】
    // 作用：瞬时 1000 个请求过来，DoEx 保证只有一个去查后端，其它人原地坐等。
    // fresh 为 true 代表你是去干活的那个人；false 代表你是白嫖结果的人。
    val, fresh, err := c.barrier.DoEx(key, func() (interface{}, error) {
   
       // 【2. 二重检查 - 读 Redis】
       if err := c.doGetCache(ctx, key, v); err != nil {
      
          // 【3. 防穿透 - 占位符判断】
          // 如果 Redis 存的是 "*"，说明 DB 也没有，直接报错返回，保护 DB。
          if err == errPlaceholder {
             return nil, c.errNotFound
          } else if err != c.errNotFound {
             // 【4. 熔断保护 - Fail Fast】
             // Redis 报错了（如超时），直接返回，不许查 DB，防止 DB 被突发流量打死。
             return nil, err
          }

          // 【5. 回源查询 - 查 MySQL】
          // 执行你在 Model 层定义的那个 select 语句。
          if err = query(v); err == c.errNotFound {
             // 【6. 空缓存写入】
             // DB 没查到，存个占位符到 Redis，有效期很短。
             if err = c.setCacheWithNotFound(ctx, key); err != nil {
                logx.WithContext(ctx).Error(err)
             }
             return nil, c.errNotFound
          } else if err != nil {
             c.stat.IncrementDbFails() // 统计 DB 失败
             return nil, err
          }

          // 【7. 数据回填 - 写 Redis】
          // 只有 DB 查到了才写缓存。
          if err = cacheVal(v); err != nil {
             logx.WithContext(ctx).Error(err)
          }
       }

       // 【8. 并发安全保护 - 序列化】
       // 将结果转为 []byte，防止多个协程操作同一个对象指针导致数据错乱。
       return jsonx.Marshal(v)
    })

    if err != nil {
       return err
    }
  
    // 如果你是干活的人，v 已经填好了，直接走
    if fresh {
       return nil
    }

    // 【9. 命中统计与数据反序列化】
    // 如果你是坐等的人，现在从排队结果里拿走属于你的那份副本。
    c.stat.IncrementTotal() // 总请求+1
    c.stat.IncrementHit()   // 命中率+1（因为你没动用 DB）

    return jsonx.Unmarshal(val.([]byte), v)
}
```

---

2. #### 三大核心思想
3. **Fail-Fast (快速失败)**：

   代码中 Redis 报错后不走数据库，体现了“宁可本次请求失败，也要保住核心数据库”的防御式编程思想。
4. **SingleFlight (请求合并)**：

   它是解决“缓存击穿”的终极方案。在单机环境下，它将 O(N) 的数据库压力瞬间降为 O(1)。
5. **Anti-Penetration (防穿透)**：

   通过 `errPlaceholder`（占位符）机制，让非法或不存在的 ID 攻击止步于 Redis 层，维持了系统的稳定性。

---

**这份笔记能帮你建立起对 `go-zero` 存储层的全局认知。要不要我帮你把这套逻辑整理成一个“高并发系统优化 Checklist”？这样你以后做其它项目优化时，直接对着清单检查就行。**

### 2.增加本地缓存

增加本地缓存，避免去redis或者mysql读数据。

API层本地缓存

```go
svc
type ServiceContext struct {
	Config     config.Config
	ProductRpc productclient.Product
	LocalCache *collection.Cache
}

func NewServiceContext(c config.Config) *ServiceContext {
	localCache, _ := collection.NewCache(2 * time.Minute)
	return &ServiceContext{
		Config:     c,
		ProductRpc: productclient.NewProduct(zrpc.MustNewClient(c.ProductRpc)),
		LocalCache: localCache,
	}
}
logic
cacheKey := fmt.Sprintf("product:id:%d", req.Id)

	if val, ok := l.svcCtx.LocalCache.Get(cacheKey); ok {
		return val.(*types.DetailResponse), nil
	}
	res, err := l.svcCtx.ProductRpc.Detail(l.ctx, &product.DetailRequest{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}

	resp := &types.DetailResponse{
		Id:     res.Id,
		Name:   res.Name,
		Desc:   res.Desc,
		Stock:  res.Stock,
		Amount: res.Amount,
		Status: res.Status,
	}
	l.svcCtx.LocalCache.Set(cacheKey, resp)
```

API层同上。

### 3.数据库缓存一致性

**`go-zero` 原生实现了“数据库与 Redis 缓存”的同步，但没有“全自动”实现“分布式本地缓存（Local Cache）”的跨节点同步。**

redis与mysql一致性源码：

```go
func (m *defaultProductModel) Insert(ctx context.Context, data *Product) (sql.Result, error) {
	productIdKey := fmt.Sprintf("%s%v", cacheProductIdPrefix, data.Id)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?)", m.table, productRowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, data.Name, data.Desc, data.Stock, data.Amount, data.Status)
	}, productIdKey)
	return ret, err
}
m.ExecCtx：
func (cc CachedConn) ExecCtx(ctx context.Context, exec ExecCtxFn, keys ...string) (
	sql.Result, error) {
	//更改数据库
  res, err := exec(ctx, cc.db)
	if err != nil {
		return nil, err
	}
	//删除缓存
	return res, cc.DelCacheCtx(ctx, keys...)
}
```

本地缓存与数据库一致性

短TTL与Pub/Sub机制实现本地缓存与数据库一致性

### 4.压力测试

#### 1.指标


| **指标名称** | **英文/缩写**   | **生活化类比**   | **核心定义**                                   |
| ------------ | --------------- | ---------------- | ---------------------------------------------- |
| **并发数**   | **Concurrency** | **收银台窗口数** | 此时此刻，系统里有多少个活跃的连接在“占位”。 |
| **吞吐量**   | **QPS**         | **每秒结账人数** | 每秒钟系统能完整处理掉多少个请求。             |
| **响应时间** | **Latency**     | **排队等候时长** | 一个请求从发起到结束，用户平均等了多久。       |

---

#### 2.核心公式：利特尔法则 (Little's Law)

这是后端性能调优的数学基石，你必须刻在脑子里：并发数=吞吐量*响应时间

- **出货量 (QPS)** 取决于：你有多少个窗口（并发） $\div$ 每个窗口干活的速度（延迟）。
- **优化的真谛**：要么多开窗口（堆服务器），要么让每个窗口干活更快（优化代码，如加本地缓存）。

---

#### 3.系统状态

随着你不断加大压力（增加并发），系统会经历三个阶段：

1. **舒适区 (Comfort Zone)**：
   - **表现**：并发增加，QPS 跟着涨，延迟基本不动。
   - **状态**：CPU 还没出汗，网卡很轻松。
2. **膝部/饱和点 (Knee Point)**：
   - **表现**：QPS 涨不动了，延迟开始向上抬头。
   - **你目前的状态**：400 并发、1.4w QPS、90ms 延迟。你正站在这个拐点上，Redis 的网络 IO 成了你的“拖油瓶”。
3. **崩溃区 (Cliff Point)**：
   - **表现**：并发继续加，QPS 掉头向下，延迟瞬间爆表，报错（Socket errors）满天飞。
   - **状态**：系统发生了“踩踏事故”，CPU 忙着上下文切换，已经不干正事了。

#### 4.实战总结：引入本地缓存

通过你的实验，我们可以得出一个极具商业价值的结论：

- **Redis 方案 (IO 密集型)**：受限于网络往返和序列化。每次请求都要在网线上飞几个来回，延迟（90ms）限制了 QPS 的上限。
- **本地缓存方案 (计算密集型)**：直接在内存里“截胡”。延迟从 **90ms** 降到 **5ms**。
- **降维打击**：根据公式，延迟降低 18 倍，意味着同样的 400 个并发连接，QPS 理论上能提升 **18 倍**！而且还省下了大量的 Redis 带宽费。

---

#### 5.压测结果

1. **报错即上限**：一旦看到 `Socket errors`（比如你之前的 400 多个），说明系统已经“冒烟”了，此时的 QPS 就是真实的极限。
2. **长尾效应 (P99)**：不要只看平均延迟（Avg）。你的 Max 延迟是 1.82s，这说明系统有严重的抖动，可能是连接池不够用或 GC 导致的。
3. **数据热点**：压测通常测的是“热点数据”。如果真实的业务是“冷数据”，性能会因为数据库磁盘 IO 而大幅下降。

未在API层增加本地缓存：

```markdown
 12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    99.61ms  177.05ms   1.94s    91.29%
    Req/Sec     1.09k   590.38     4.81k    67.25%
  388105 requests in 30.10s, 101.78MB read
  Socket errors: connect 0, read 495, write 0, timeout 39
Requests/sec:  12893.58
Transfer/sec:      3.38MB
```

在API层增加本地缓存：

```markdown
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    79.13ms  138.16ms   1.41s    92.69%
    Req/Sec     2.39k     1.25k    9.52k    69.79%
  856361 requests in 30.07s, 224.59MB read
  Socket errors: connect 0, read 581, write 0, timeout 7
Requests/sec:  28480.79
Transfer/sec:      7.47MB
```

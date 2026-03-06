# drugo-provider

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

基于 `drugo` 框架的服务提供者包，提供了数据库、Redis 缓存、国际化和 Gin Web 服务的集成支持。

## 🚀 特性

- **数据库服务** (`dbsvc`): 基于 `mgorm` 的多数据库连接管理
- **Redis 服务** (`redissvc`): 基于 `mgredis` 的 Redis 缓存管理  
- **国际化服务** (`i18nsvc`): 基于 `mi18n` 的多语言翻译支持
- **Gin 服务** (`ginsrv`): 基于 `gin-gonic` 的 Web 框架集成
- **便捷函数** (`pkg/svc`): 简化数据库、Redis和i18n服务获取的语义化封装
- **配置驱动**: 通过配置文件灵活管理各种服务
- **优雅关闭**: 支持服务的优雅启动和关闭
- **日志集成**: 内置 `zap` 日志支持

## 📦 安装

```bash
go get github.com/qq1060656096/drugo-provider
```

## 🔧 快速开始

### 基本使用

#### 使用便捷函数（推荐）

```go
import(
    "github.com/qq1060656096/drugo-provider/pkg/svc"
)

// 获取默认数据库连接
db := svc.MustDefaultDB(c)
companyInfo := make(map[string]interface{})
db.Raw("select * from common_company where company_id= 218908").Scan(&companyInfo)

// 获取会话 Redis 客户端
sessionRedis := svc.MustSessionRedis(c)
r, err := sessionRedis.Set(c.Request.Context(), "api", "demo", 0).Result()

// 获取国际化服务并翻译
welcome := svc.MustT(c, "zh", "welcome", nil) // 输出: 欢迎
data := map[string]any{"Name": "张三"}
greeting := svc.MustT(c, "zh", "greeting", data) // 输出: 你好，张三！
```

#### 使用原始服务

```go
import(
    "github.com/qq1060656096/drugo-provider/dbsvc"
    "github.com/qq1060656096/drugo-provider/ginsrv"
    "github.com/qq1060656096/drugo-provider/redissvc"
    "github.com/qq1060656096/drugo/drugo"
)

// 获取数据库 service
dbSvc := ginsrv.MustGetService[*drugo.Drugo, *dbsvc.DbService](c, dbsvc.Name)
db := dbSvc.Manager().MustGroup("public").MustGet(c.Request.Context(), "test_common")
companyInfo := make(map[string]interface{})
db.Raw("select * from common_company where company_id= 218908").Scan(&companyInfo)

// 获取 redis 缓存
redisSvc := ginsrv.MustGetService[*drugo.Drugo, *redissvc.RedisService](c, redissvc.Name)
r, err := redisSvc.Group().MustGet(c, "session").Set(c.Request.Context(), "api", "demo", 0).Result()
```

### 服务注册

```go
import (
    "github.com/qq1060656096/drugo/drugo"
    "github.com/qq1060656096/drugo-provider/dbsvc"
    "github.com/qq1060656096/drugo-provider/ginsrv"
    "github.com/qq1060656096/drugo-provider/i18nsvc"
    "github.com/qq1060656096/drugo-provider/redissvc"
)

func main() {
    app := drugo.New()
    
    // 注册服务
    app.Register(dbsvc.New())
    app.Register(redissvc.New())
    app.Register(i18nsvc.New())
    app.Register(ginsrv.New())
    
    // 启动应用
    if err := app.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

## 📋 服务说明

### 数据库服务 (dbsvc)

提供基于 `mgorm` 的数据库连接管理，支持多种数据库类型：

- MySQL
- PostgreSQL  
- SQLite
- SQL Server

**配置示例**:

```yaml
db:
  public:
    test_common:
      driver_type: mysql
      dsn: "user:password@tcp(localhost:3306)/database?charset=utf8mb4&parseTime=True&loc=Local"
      max_idle_conns: 10
      max_open_conns: 100
      conn_max_lifetime: 1h
```

### 国际化服务 (i18nsvc)

提供基于 `mi18n` 的多语言翻译支持，支持多种文件格式和模板变量。

**配置示例**:

```yaml
i18n:
  locale_dir: "locale"    # 翻译文件目录
  default_lang: "en"       # 默认语言
```

**翻译文件示例**:

locale/zh.json
```json
[
  { "id": "welcome", "translation": "欢迎" },
  { "id": "greeting", "translation": "你好，{{.Name}}！" }
]
```

locale/en.json
```json
[
  { "id": "welcome", "translation": "Welcome" },
  { "id": "greeting", "translation": "Hello, {{.Name}}!" }
]
```

### Redis 服务 (redissvc)

提供基于 `mgredis` 的 Redis 连接管理，支持连接池配置。

**配置示例**:

```yaml
redis:
  session:
    addr: "localhost:6379"
    password: ""
    db: 0
    pool_size: 10
    min_idle_conns: 5
    dial_timeout: 5s
    read_timeout: 3s
    write_timeout: 3s
  cache:
    addr: "localhost:6379"
    password: ""
    db: 1
    pool_size: 20
```

### Gin 服务 (ginsrv)

提供基于 `gin-gonic` 的 Web 框架集成，支持 HTTP/HTTPS 双协议。

**配置示例**:

```yaml
gin:
  mode: "release"  # debug, release, test
  host: "0.0.0.0"
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s
  shutdown_timeout: 30s
  http:
    enabled: true
    port: 8080
  https:
    enabled: false
    port: 8443
    cert_file: "cert.pem"
    key_file: "key.pem"
```

**中间件使用示例**：

说明：建议先使用 `ginSvc.SetEngineContextAppVar(app)` 将 `app` 注入到 `gin.Context`，这样在后续 handler 中才能通过 `pkg/svc` 或 `ginsrv.MustGetService` 从请求上下文获取服务。

```go
import (
    "context"

    "github.com/qq1060656096/drugo/drugo"
    "github.com/qq1060656096/drugo-provider/ginsrv"
)

func main() {
    ctx := context.Background()

    app := drugo.New()

    app.Register(ginsrv.New())

    ginSvc := drugo.MustGetService[*ginsrv.GinService](app, ginsrv.Name)

    ginSvc.SetEngineContextAppVar(app)

    engine := ginSvc.Engine()

    engine.Use(ginsrv.TraceMiddleware("X-Request-ID"))
    engine.Use(ginsrv.AccessLoggerWithoutBody(app.Logger(), "gin.access", "gin.error"))
    engine.Use(ginsrv.RecoveryLogger(app.Logger(), "gin.error"))

    if err := app.Run(ctx); err != nil {
        panic(err)
    }
}
```

## 🛠️ API 文档

### 便捷服务函数 (pkg/svc)

为了简化开发，`pkg/svc` 包提供了便捷的函数来快速获取数据库、Redis和i18n服务：

#### 数据库便捷函数

```go
import "github.com/qq1060656096/drugo-provider/pkg/svc"

// 获取数据库服务实例
dbSvc := svc.MustDB(c)

// 获取默认数据库连接
db := svc.MustDefaultDB(c)
db.Find(&users)

// 获取公共库连接
publicDB := svc.MustPublicDB(c)
publicDB.Raw("SELECT * FROM common_config").Scan(&configs)

// 获取指定业务数据库连接
businessDB := svc.MustBusinessDB(c, "data_1")
businessDB.Create(&businessData)
```

#### Redis 便捷函数

```go
import "github.com/qq1060656096/drugo-provider/pkg/svc"

// 获取 Redis 服务实例
redisSvc := svc.MustRedis(c)

// 获取默认 Redis 客户端
redisClient := svc.MustDefaultRedis(c)
redisClient.Set(ctx, "key", "value", time.Hour)

// 获取购物车 Redis 客户端
cartRedis := svc.MustCartRedis(c)
cartRedis.LPush(ctx, "cart:123", "item1")

// 获取会话 Redis 客户端
sessionRedis := svc.MustSessionRedis(c)
sessionRedis.Set(ctx, "session:abc", "userdata", 30*time.Minute)
```

#### 国际化便捷函数

```go
import "github.com/qq1060656096/drugo-provider/pkg/svc"

// 获取i18n服务实例
i18nSvc := svc.MustI18n(c)

// 基本翻译
welcome := svc.MustT(c, "zh", "welcome", nil) // 输出: 欢迎

// 带变量的翻译
data := map[string]any{"Name": "张三"}
greeting := svc.MustT(c, "zh", "greeting", data) // 输出: 你好，张三！

// 使用Context翻译
ctxWithLang := svc.MustWithLang(c, "zh")
c.Request = c.Request.WithContext(ctxWithLang)
message := svc.MustTCtx(c, "welcome", nil) // 输出: 欢迎

// 获取当前语言
lang := svc.MustLang(c) // 输出: zh

// 获取支持的语言列表
languages := svc.MustGetSupportedLanguages(c)
fmt.Printf("支持的语言: %v", languages)

// 重新加载翻译文件（在翻译文件更新后）
if err := svc.MustReloadI18n(c); err != nil {
    log.Printf("重新加载失败: %v", err)
}
```

### 数据库服务 API

```go
// 获取数据库服务
dbSvc := ginsrv.MustGetService[*drugo.Drugo, *dbsvc.DbService](c, dbsvc.Name)

// 获取数据库管理器
manager := dbSvc.Manager()

// 获取指定分组的数据库
db := manager.MustGroup("public").MustGet(ctx, "test_common")

// 执行 SQL 查询
var results []Model
db.Raw("SELECT * FROM table WHERE id = ?", id).Scan(&results)

// 使用 GORM 操作
db.Create(&model)
db.Find(&models)
db.Where("id = ?", id).First(&model)
```

### Redis 服务 API

```go
// 获取 Redis 服务
redisSvc := ginsrv.MustGetService[*drugo.Drugo, *redissvc.RedisService](c, redissvc.Name)

// 获取 Redis 客户端
client := redisSvc.Group().MustGet(c, "session")

// 字符串操作
client.Set(ctx, "key", "value", time.Hour)
val, err := client.Get(ctx, "key").Result()

// 哈希操作
client.HSet(ctx, "hash_key", "field", "value")
val, err = client.HGet(ctx, "hash_key", "field").Result()

// 列表操作
client.LPush(ctx, "list_key", "item1", "item2")
vals, err := client.LRange(ctx, "list_key", 0, -1).Result()
```

### Gin 服务 API

```go
// 获取 Gin 服务
ginSvc := ginsrv.MustGetService[*drugo.Drugo, *ginsrv.GinService](c, ginsrv.Name)

// 获取 Gin 引擎
engine := ginSvc.Engine()

// 注册路由
engine.GET("/users", func(c *gin.Context) {
    // 处理逻辑
})

engine.POST("/users", func(c *gin.Context) {
    // 处理逻辑
})

// 中间件
engine.Use(gin.Logger())
engine.Use(gin.Recovery())
```

## 🧪 测试

运行测试：

```bash
go test ./...
```

运行特定包的测试：

```bash
go test ./dbsvc
go test ./redissvc  
go test ./ginsrv
go test ./i18nsvc
```

## 📄 许可证

本项目采用 Apache 许可证。详见 [LICENSE](LICENSE) 文件。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 🔗 相关链接

- [drugo 框架](https://github.com/qq1060656096/drugo)
- [mgorm](https://github.com/qq1060656096/mgorm)
- [mgredis](https://github.com/qq1060656096/mgredis)
- [mi18n](https://github.com/qq1060656096/mi18n)
- [Gin Web 框架](https://github.com/gin-gonic/gin)
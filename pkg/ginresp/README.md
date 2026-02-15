# ginresp

Gin 场景下的统一 JSON 响应工具包，提供极简 API 和标准化响应格式。

## 设计目标

- **极简 API** - 无需配置，开箱即用
- **标准库风格** - 简洁直观的函数命名
- **单一输出路径** - 统一的 JSON 响应格式
- **自动错误处理** - error -> HTTP status 自动映射
- **自动追踪** - 自动添加 trace ID 到响应

## 安装

```go
import "github.com/qq1060656096/drugo-provider/pkg/ginresp"
```

## 快速开始

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/qq1060656096/drugo-provider/pkg/ginresp"
)

func main() {
    r := gin.Default()
    
    r.GET("/success", func(c *gin.Context) {
        ginresp.OK(c, map[string]string{"message": "success"})
    })
    
    r.GET("/error", func(c *gin.Context) {
        err := errors.New("something went wrong")
        ginresp.Err(c, err)
    })
    
    r.Run()
}
```

## API 文档

### 成功响应

#### `OK(c *gin.Context, data any)`
返回成功响应（HTTP 200）。

```go
ginresp.OK(c, "success")
ginresp.OK(c, map[string]string{"key": "value"})
ginresp.OK(c, nil) // 返回空数据
```

#### `OKMsg(c *gin.Context, data any, msg string)`
返回带自定义消息的成功响应。

```go
ginresp.OKMsg(c, data, "操作成功")
```

### 错误响应

#### `Fail(c *gin.Context, code int, msg string)`
返回业务错误（固定 HTTP 200，适合前端业务码判断）。

```go
ginresp.Fail(c, 1001, "参数无效")
ginresp.Fail(c, 2001, "用户不存在")
```

#### `Err(c *gin.Context, err error)`
根据 error 自动生成响应，支持 `errcode.Error` 类型。

```go
// 普通错误 - 返回 HTTP 500
ginresp.Err(c, errors.New("internal error"))

// errcode.Error - 自动映射 HTTP 状态码
err := errcode.New(1014000001, "参数错误") // HTTP 400
ginresp.Err(c, err)
```

### 终止链式响应

以下函数在发送响应后会调用 `c.Abort()` 终止中间件链：

#### `AbortOK(c *gin.Context, data any)`
返回成功并终止链。

#### `AbortFail(c *gin.Context, code int, msg string)`
返回业务错误并终止链。

#### `AbortErr(c *gin.Context, err error)`
返回错误并终止链。

## 响应格式

### 成功响应
```json
{
  "code": 0,
  "message": "OK",
  "data": {...},
  "trace_id": "abc123" // 可选，如果设置了 trace ID
}
```

### 业务错误响应
```json
{
  "code": 1001,
  "message": "参数无效",
  "data": null,
  "trace_id": "abc123" // 可选
}
```

### 系统错误响应
```json
{
  "code": 1500000001,
  "message": "internal error",
  "data": null,
  "trace_id": "abc123" // 可选
}
```

## Trace ID 支持

包会自动从 Gin Context 中获取 trace ID 并添加到响应中。

```go
// 在中间件中设置 trace ID
func TraceMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        traceID := generateTraceID()
        c.Set(ginresp.TraceIDKey, traceID)
        c.Next()
    }
}
```

## 错误码规范

推荐使用 9 位错误码格式：
```
1(占位符) + 01(模块) + 400(HTTP状态) + 0001(顺序)
```

例如：
- `1014000001` - 参数错误（HTTP 400）
- `1015000001` - 内部错误（HTTP 500）

## 性能

包经过性能优化，基准测试结果：

```
BenchmarkOK-8         1000000    1200 ns/op    1024 B/op    5 allocs/op
BenchmarkFail-8       1000000    1150 ns/op     960 B/op    4 allocs/op
BenchmarkErr-8        1000000    1180 ns/op     992 B/op    5 allocs/op
```

## 依赖

- `github.com/gin-gonic/gin`
- `github.com/qq1060656096/bizutil/eresp`
- `github.com/qq1060656096/bizutil/errcode`

## 测试

运行测试：

```bash
go test ./pkg/ginresp/...
```

运行基准测试：

```bash
go test -bench=. ./pkg/ginresp/...
```

## 许可证

本项目采用 MIT 许可证。

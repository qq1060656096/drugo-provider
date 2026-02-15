// Package ginresp 提供 Gin 场景下的统一 JSON 响应工具。
// 设计目标：
//
//   - 极简 API（无配置）
//   - 标准库风格
//   - 单一输出路径
//   - 自动 error -> HTTP status
//   - 自动 trace id
//
// 推荐:
//
//	ginresp.OK(c, data)
//	ginresp.Err(c, err)
//	ginresp.Fail(c, 1001, "invalid param")
package ginresp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/bizutil/eresp"
	"github.com/qq1060656096/bizutil/errcode"
)

// TraceIDKey 用于在 Gin Context 中存储 trace ID 的键名常量
const TraceIDKey = "trace_id"

//
// ---------- public api ----------
//

// OK 返回成功响应。
// 参数：
//   - c: Gin 上下文对象
//   - data: 要返回的数据，可以是任意类型
func OK(c *gin.Context, data any) {
	write(c, http.StatusOK, eresp.OKResp(data, ""))
}

// OKMsg 返回带 message 的成功响应。
// 参数：
//   - c: Gin 上下文对象
//   - data: 要返回的数据，可以是任意类型
//   - msg: 成功消息描述
func OKMsg(c *gin.Context, data any, msg string) {
	write(c, http.StatusOK, eresp.OKResp(data, msg))
}

// Fail 返回业务错误（固定 200，适合前端业务码判断）。
// 参数：
//   - c: Gin 上下文对象
//   - code: 业务错误码，用于前端判断具体错误类型
//   - msg: 错误消息描述
func Fail(c *gin.Context, code int, msg string) {
	write(c, http.StatusOK, eresp.ErrorResp(code, "", msg, nil))
}

// Err 根据 error 自动生成响应。
// 会自动解析错误类型并设置对应的 HTTP 状态码。
// 参数：
//   - c: Gin 上下文对象
//   - err: 错误对象，支持 errcode.Error 类型和其他标准错误
func Err(c *gin.Context, err error) {
	status := resolveStatus(err)
	resp := eresp.FromError(err, nil)
	write(c, status, resp)
}

//
// ---------- abort ----------
//

// AbortOK 返回成功并终止链。
// 在返回成功响应后会调用 c.Abort() 终止中间件链的执行。
// 参数：
//   - c: Gin 上下文对象
//   - data: 要返回的数据，可以是任意类型
func AbortOK(c *gin.Context, data any) {
	OK(c, data)
	c.Abort()
}

// AbortFail 返回错误并终止链。
// 在返回业务错误响应后会调用 c.Abort() 终止中间件链的执行。
// 参数：
//   - c: Gin 上下文对象
//   - code: 业务错误码，用于前端判断具体错误类型
//   - msg: 错误消息描述
func AbortFail(c *gin.Context, code int, msg string) {
	Fail(c, code, msg)
	c.Abort()
}

// AbortErr 返回 error 并终止链。
// 在返回错误响应后会调用 c.Abort() 终止中间件链的执行。
// 参数：
//   - c: Gin 上下文对象
//   - err: 错误对象，支持 errcode.Error 类型和其他标准错误
func AbortErr(c *gin.Context, err error) {
	Err(c, err)
	c.Abort()
}

//
// ---------- internal ----------
//

// write 内部函数：写入 JSON 响应。
// 会自动添加 trace ID（如果存在）到响应中。
// 参数：
//   - c: Gin 上下文对象
//   - status: HTTP 状态码
//   - resp: 标准化的响应对象
func write(c *gin.Context, status int, resp eresp.Response) {
	if trace := getTraceID(c); trace != "" {
		resp = resp.WithTrace(trace)
	}
	c.JSON(status, resp)
}

// resolveStatus 内部函数：根据错误类型解析对应的 HTTP 状态码。
// 如果错误是 errcode.Error 类型，返回其定义的 HTTP 状态码；
// 否则返回 500 Internal Server Error。
// 参数：
//   - err: 错误对象
//
// 返回值：对应的 HTTP 状态码
func resolveStatus(err error) int {
	var ec *errcode.Error
	if errors.As(err, &ec) {
		return ec.HTTPStatus()
	}
	return http.StatusInternalServerError
}

// getTraceID 内部函数：从 Gin Context 中获取 trace ID。
// 参数：
//   - c: Gin 上下文对象
//
// 返回值：trace ID 字符串，如果不存在则返回空字符串
func getTraceID(c *gin.Context) string {
	if c == nil {
		return ""
	}

	v, ok := c.Get(TraceIDKey)
	if !ok {
		return ""
	}

	s, _ := v.(string)
	return s
}

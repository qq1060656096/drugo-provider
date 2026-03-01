package ginsrv

import (
	"net"

	"github.com/gin-gonic/gin"
)

// 辅助函数：获取客户端IP
func getClientIP(c *gin.Context) string {
	clientIP := c.ClientIP()
	if clientIP == "" {
		ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err == nil {
			return ip
		}
	}
	return clientIP
}

// GetVar 从 gin.Context 中获取指定 key 的值并尝试转换为指定类型 T。
//
// 如果 key 不存在，或存在但类型与 T 不匹配，则返回 T 的零值，exists 为 false。
//
// 该函数不会 panic，适用于安全读取场景。
//
// 示例：
//
//	user, ok := GetVar[*User](c, "user")
//	if !ok {
//	    return errors.New("user not found")
//	}
//
// 参数：
//   - c: gin 请求上下文
//   - key: 存储在 Context 中的键名
//
// 返回：
//   - v: 转换后的值（失败时为 T 的零值）
//   - exists: 是否存在且类型匹配
func GetVar[T any](c *gin.Context, key string) (v T, exists bool) {
	raw, ok := c.Get(key)
	if !ok {
		return v, false
	}

	v, ok = raw.(T)
	return v, ok
}

// MustGetVar 从 gin.Context 中获取指定 key 的值并转换为指定类型 T。
//
// 如果 key 不存在，或值类型与 T 不匹配，该函数会 panic。
//
// 该函数适用于确定值一定存在的场景（如认证 middleware 已保证 user 存在）。
//
// 行为等价于 gin.Context.MustGet + 类型断言，但提供更安全的类型检查。
//
// 示例：
//
//	user := MustGetVar[*User](c, "user")
//
// 参数：
//   - c: gin 请求上下文
//   - key: 存储在 Context 中的键名
//
// 返回：
//   - 转换后的值（类型为 T）
//
// panic：
//   - key 不存在
//   - 类型不匹配
func MustGetVar[T any](c *gin.Context, key string) T {
	raw := c.MustGet(key)

	v, ok := raw.(T)
	if !ok {
		panic("gin: context value type mismatch for key: " + key)
	}

	return v
}

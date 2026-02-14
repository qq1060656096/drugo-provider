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

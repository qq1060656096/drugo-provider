package svc

import (
	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/drugo-provider/ginsrv"
	"github.com/qq1060656096/drugo-provider/pkg/consts"
	"github.com/qq1060656096/drugo-provider/redissvc"
	"github.com/qq1060656096/drugo/drugo"
	"github.com/redis/go-redis/v9"
)

// MustRedis 从 gin.Context 中获取 RedisService 实例。
//
// 该方法基于 ginsrv.MustGetService：
// - 从 Gin Context 中解析 *drugo.Drugo 实例
// - 根据 redissvc.Name 查找并返回对应的 RedisService
//
// 如果 RedisService 未注册或获取失败，将直接 panic，
// 适用于：
//   - 框架初始化阶段
//   - 中间件 / Handler 中“必须存在 Redis”的场景
func MustRedis(ctx *gin.Context) *redissvc.RedisService {
	return ginsrv.MustGetService[*drugo.Drugo, *redissvc.RedisService](
		ctx,
		redissvc.Name,
	)
}

// MustDefaultRedis 返回默认 Redis Client。
//
// 该方法从 gin.Context 中获取 RedisService，
// 并基于当前 HTTP 请求的 Context 获取 default Redis 实例。
//
// 若 RedisService 未注册或 default Redis 不存在，
// 将直接 panic。
func MustDefaultRedis(ctx *gin.Context) *redis.Client {
	redisSvc := MustRedis(ctx)
	return redisSvc.Group().MustGet(ctx.Request.Context(), consts.RedisDefault)
}

// MustCartRedis 从 Gin 上下文中获取购物车 Redis 客户端。
//
// 该方法基于当前请求上下文（Request Context），
// 从 Redis 服务组中获取 cart 业务使用的 Redis 实例。
//
// 若 Redis 服务未初始化、分组不存在，或获取过程中发生错误，
// 将直接 panic，用于必须依赖 Redis 的业务场景。
func MustCartRedis(ctx *gin.Context) *redis.Client {
	redisSvc := MustRedis(ctx)
	return redisSvc.Group().MustGet(ctx.Request.Context(), consts.RedisCart)
}

// MustSessionRedis 从 Gin 上下文中获取会话（Session）Redis 客户端。
//
// 该方法基于当前请求上下文（Request Context），
// 从 Redis 服务组中获取 session 业务使用的 Redis 实例。
//
// 若 Redis 服务未初始化、分组不存在，或获取过程中发生错误，
// 将直接 panic，用于必须依赖 Redis 的中间件或核心业务流程。
func MustSessionRedis(ctx *gin.Context) *redis.Client {
	redisSvc := MustRedis(ctx)
	return redisSvc.Group().MustGet(ctx.Request.Context(), consts.RedisSession)
}

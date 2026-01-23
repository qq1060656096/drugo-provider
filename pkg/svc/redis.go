package svc

import (
	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/drugo-provider/ginsrv"
	"github.com/qq1060656096/drugo-provider/redissvc"
	"github.com/qq1060656096/drugo/drugo"
)

func MustRedis(ctx *gin.Context) *redissvc.RedisService {
	return ginsrv.MustGetService[*drugo.Drugo, *redissvc.RedisService](
		ctx,
		redissvc.Name,
	)
}

package svc

import (
	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/drugo-provider/dbsvc"
	"github.com/qq1060656096/drugo-provider/ginsrv"
	"github.com/qq1060656096/drugo/drugo"
)

func MustDB(ctx *gin.Context) *dbsvc.DbService {
	return ginsrv.MustGetService[*drugo.Drugo, *dbsvc.DbService](
		ctx,
		dbsvc.Name,
	)
}

package api

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/drugo-provider/biapi/biz"
	"github.com/qq1060656096/drugo-provider/biapi/data"
	"github.com/qq1060656096/drugo-provider/biapi/service"
	"github.com/qq1060656096/drugo-provider/dbsvc"
	"github.com/qq1060656096/drugo/drugo"
	"github.com/qq1060656096/drugo/pkg/router"
	"go.uber.org/zap"
)

var defaultHandler *BiHandler
var defaultGroupName = "bi"
var defaultDbName = "bi_data"
var once sync.Once

func Init(groupName string, dbName string) {
	once.Do(func() {
		defaultGroupName = groupName
		defaultDbName = dbName
		router.Default().Register(func(engine *gin.Engine) {
			defaultHandler = NewBiHandler(defaultGroupName, defaultDbName)
			defaultHandler.RegisterRoutes(engine)
		})
	})
}

type BiHandler struct {
	service   *service.BiService
	groupName string
	dbName    string
}

func NewBiHandler(groupName string, dbName string) *BiHandler {
	repo := data.NewBiRepo()
	uc := biz.NewBiUsecase(repo)
	service := service.NewBiService(uc)
	return &BiHandler{
		service:   service,
		groupName: groupName,
		dbName:    dbName,
	}
}
func (h *BiHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/api/bi/v1/debug/:code", h.Execute)
	router.POST("/api/bi/v1/:code", h.Execute)

}

func (h *BiHandler) Execute(ctx *gin.Context) {
	req := &service.ExecuteRequest{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Code = ctx.Param("code")
	if req.Code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
	}

	dbService := drugo.MustGetService[*dbsvc.DbService](drugo.App(), "db")
	appLogger := drugo.App().Logger()

	tplDb := dbService.Manager().MustGroup(h.groupName).MustGet(ctx, h.dbName)
	execDb := dbService.Manager().MustGroup(h.groupName).MustGet(ctx, h.dbName)
	result, err := h.service.Execute(ctx, tplDb, execDb, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	appLogger.MustGet("bi").Info("", zap.Any("result", result))

	ctx.JSON(http.StatusOK, result)
}

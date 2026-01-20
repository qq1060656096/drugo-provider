package data

import (
	"context"

	"github.com/qq1060656096/bizutil/qsql"
	"github.com/qq1060656096/drugo-provider/biapi/biz"
	"github.com/qq1060656096/drugo/drugo"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

const Name = "bi"

var _ biz.BiRepo = (*BiRepo)(nil)

type BiRepo struct {
	tplRepo *templateRepo
	name    string
}

func (b *BiRepo) Execute(ctx context.Context, tplDb, execDB *gorm.DB, req *biz.ExecuteRequest) (*biz.ExecuteResult, error) {
	buildResult, err := b.Build(ctx, tplDb, req)
	appLogger := drugo.App().Logger().MustGet(Name)
	if err != nil {
		appLogger.Error("BiRepo.Build", zap.Error(err),
			zap.Any("req", req),
			zap.Any("buildResult", buildResult),
		)
		return nil, err
	}
	db := execDB.WithContext(ctx)
	var returnData any
	var count int64
	var rowsAffected int64
	sql := buildResult.SQLStmt.SQL
	args := buildResult.SQLStmt.Args
	switch buildResult.OpType {
	case biz.OpTypeList:
		var data []map[string]any
		err := db.Raw(sql, args...).Scan(&data).Error
		if err != nil {
			return nil, err
		}
		returnData = data
		rowsAffected = int64(len(data))

	case biz.OpTypeDetail:
		var detail map[string]any
		err := db.Raw(sql, args...).Scan(&detail).Error
		if err != nil {
			return nil, err
		}
		returnData = detail
		rowsAffected = 1
	case biz.OpTypeCount:
		err := db.Raw(sql, args...).Scan(&count).Error
		if err != nil {
			return nil, err
		}
		rowsAffected = count
	case biz.OpTypeAdd, biz.OpTypeUpdate, biz.OpTypeDel:
		result := db.Exec(sql, args...)
		if result.Error != nil {
			return nil, result.Error
		}
		rowsAffected = result.RowsAffected
	}

	executeResult := &biz.ExecuteResult{
		OpType:           buildResult.OpType,
		RowsAffected:     rowsAffected,
		Data:             returnData,
		Count:            count,
		ValidatorsErrors: buildResult.SQLStmt.ValidatorsErrors,
		BuildResult:      buildResult,
	}
	if executeResult.Errors == nil {
		executeResult.Errors = []error{}
	}
	if executeResult.ValidatorsErrors == nil {
		executeResult.ValidatorsErrors = []*qsql.ValidatorError{}
	}
	return executeResult, nil
}

func (b *BiRepo) Build(ctx context.Context, tplDb *gorm.DB, req *biz.ExecuteRequest) (*biz.BuildResult, error) {
	tpl, err := b.tplRepo.FindTpl(ctx, tplDb, req.PlatformId, req.Code)
	appLogger := drugo.App().Logger().MustGet(Name)
	if err != nil {
		appLogger.Error("BiRepo.Build template not found", zap.Error(err), zap.Any("req", req))
		return nil, err
	}
	tplId := tpl.TemplateId
	tplData, err := b.tplRepo.FindTplData(ctx, tplDb, req.PlatformId, tplId, req.CompanyId, req.Env)
	if err != nil {
		appLogger.Error("BiRepo.Build template data not found", zap.Error(err), zap.Any("req", req))
		return nil, err
	}
	content := tplData.Content
	qe := qsql.NewEngine()
	err = qe.Parse("sql", content)
	if err != nil {
		appLogger.Error("BiRepo.Build template content parse", zap.Error(err), zap.Int64("tplId", tplId), zap.Any("req", req))
		return nil, err
	}
	vars := qsql.NewValueVars()
	vars.Params(req.Params)
	vars.Sys(req.Sys)
	vars.Users(req.Users)

	stm, err := qe.ExecuteWithVars(vars)
	if err != nil {
		appLogger.Error("BiRepo.Build template execution", zap.Error(err), zap.Int64("tplId", tplId), zap.Any("req", req), zap.Any("stm", stm))
		return nil, err
	}
	rt := &biz.BuildResult{
		TdId:    tplData.TdId,
		OpType:  tplData.OpType,
		SQLStmt: stm,
	}
	return rt, nil
}

func NewBiRepo() *BiRepo {
	return &BiRepo{
		tplRepo: newTemplateRepo(),
		name:    "biapi",
	}
}

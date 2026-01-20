// Package service 提供 BI 模板的服务编排层。
// 本包负责对外暴露能力，协调 Biz 层完成业务流程。
package service

import (
	"context"

	"github.com/qq1060656096/drugo-provider/biapi/biz"
	"gorm.io/gorm"
)

type ExecuteRequest struct {
	biz.ExecuteRequest
}

type ExecuteResult struct {
	biz.ExecuteResult
}

type BiService struct {
	uc *biz.BiUsecase
}

func NewBiService(uc *biz.BiUsecase) *BiService {
	return &BiService{
		uc: uc,
	}
}

// Execute 执行 BI 模板，返回生成的 SQL、参数和查询结果。
func (s *BiService) Execute(ctx context.Context, tplDb, execDB *gorm.DB, req *ExecuteRequest) (*biz.ExecuteResult, error) {
	return s.uc.Execute(ctx, tplDb, execDB, &req.ExecuteRequest)
}

// Build 仅解析 DSL 并生成 SQL，不执行查询。
func (s *BiService) Build(ctx context.Context, tplDb *gorm.DB, req *ExecuteRequest) (*biz.BuildResult, error) {
	return s.uc.Build(ctx, tplDb, &req.ExecuteRequest)
}

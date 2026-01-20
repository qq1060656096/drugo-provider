// Package biz 提供 BI 模板的业务逻辑层。
// 本包负责业务规则处理，协调 Data 层获取数据，并使用 qsql 执行 DSL。
package biz

import (
	"context"
	"errors"

	"github.com/qq1060656096/bizutil/qsql"
	"gorm.io/gorm"
)

// 操作类型（按语义分段，禁止随意调整）
const (
	// 2xx：写操作
	OpTypeAdd    = 201
	OpTypeUpdate = 202
	OpTypeDel    = 203

	// 4xx：读操作
	OpTypeList   = 401
	OpTypeDetail = 402
	OpTypeCount  = 403
)

// 预定义的环境常量。
const (
	EnvTest = "test"
	EnvGray = "gray"
	EnvProd = "prod"
)

// 预定义错误。
var (
	ErrTemplateNotFound     = errors.New("biz: template not found")
	ErrTemplateDataNotFound = errors.New("biz: template data not found")
	ErrDSLParseFailed       = errors.New("biz: dsl parse failed")
	ErrDSLExecuteFailed     = errors.New("biz: dsl execute failed")
	ErrUnsupportedOpType    = errors.New("biz: unsupported op type")
)

// ExecuteRequest 表示 BI 模板执行请求。
type ExecuteRequest struct {
	PlatformId int64  `json:"platform_id"` // 平台 ID
	CompanyId  int64  `json:"company_id"`  // 公司 ID
	Code       string `json:"code"`        // 模板业务编码
	Env        string `json:"env"`         // 环境: test/gray/prod
	Params     any    `json:"params"`      // 用户传入的查询参数
	Sys        any    `json:"sys"`         // 系统参数（如当前用户信息）
	Users      any    `json:"users"`       // 用户相关信息
	Page       int    `json:"page"`        // 页码，从 1 开始
	PageSize   int    `json:"page_size"`   // 每页数量
}

// ExecuteResult 表示 BI 模板执行结果。
type ExecuteResult struct {
	Data             any                    `json:"data"`             // 查询结果列表
	Count            int64                  `json:"count"`            // 总记录数（list/count）
	ValidatorsErrors []*qsql.ValidatorError `json:"validator_errors"` // DSL 校验错误
	Errors           []error                `json:"errors"`
	RowsAffected     int64                  `json:"rows_affected"`          // 受影响行数
	OpType           int                    `json:"op_type"`                // 操作类型
	BuildResult      *BuildResult           `json:"build_result,omitempty"` // 构建结果（调试用）
}

type BuildResult struct {
	TdId    int64
	OpType  int
	SQLStmt *qsql.SQLStmt
}

// TemplateUsecase 定义 BI 模板业务逻辑接口。
type BiRepo interface {
	// Execute 执行 BI 模板，返回生成的 SQL、参数和查询结果。
	Execute(ctx context.Context, tplDb, execDB *gorm.DB, req *ExecuteRequest) (*ExecuteResult, error)

	// Build 仅解析 DSL 并生成 SQL，不执行查询。
	Build(ctx context.Context, tplDb *gorm.DB, req *ExecuteRequest) (*BuildResult, error)
}

type BiUsecase struct {
	repo BiRepo
}

func NewBiUsecase(repo BiRepo) *BiUsecase {
	return &BiUsecase{
		repo: repo,
	}
}

// Execute 执行 BI 模板，返回生成的 SQL、参数和查询结果。
func (u *BiUsecase) Execute(ctx context.Context, tplDb, execDB *gorm.DB, req *ExecuteRequest) (*ExecuteResult, error) {
	return u.repo.Execute(ctx, tplDb, execDB, req)
}

// Build 仅解析 DSL 并生成 SQL，不执行查询。
func (u *BiUsecase) Build(ctx context.Context, tplDb *gorm.DB, req *ExecuteRequest) (*BuildResult, error) {
	return u.repo.Build(ctx, tplDb, req)
}

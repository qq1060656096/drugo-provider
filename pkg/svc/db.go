package svc

import (
	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/drugo-provider/dbsvc"
	"github.com/qq1060656096/drugo-provider/ginsrv"
	"github.com/qq1060656096/drugo-provider/pkg/consts"
	"github.com/qq1060656096/drugo/drugo"
	"github.com/qq1060656096/mgorm"
	"gorm.io/gorm"
)

// MustDB 从 Gin 上下文中获取数据库服务（DbService）。
//
// 该方法是对 ginsrv.MustGetService 的语义化封装，
// 用于在 HTTP 请求生命周期内获取当前应用绑定的数据库服务实例。
//
// 设计目的：
//   - 屏蔽底层 Service 获取细节，提升业务代码可读性
//   - 统一 DbService 的获取入口，避免在业务层直接依赖 ginsrv
//
// 使用约定：
//   - ctx 必须是当前请求对应的 *gin.Context
//   - DbService 需已在 Gin 中间件阶段完成注册
//
// 注意事项：
//   - 使用 Must 语义：
//   - 当服务未注册
//   - 或类型断言失败
//   - 或上下文不合法
//     时将直接 panic
//   - 适合在 Handler / Service / Data 层中使用，
//     不建议在可恢复错误路径中调用
func MustDB(ctx *gin.Context) *dbsvc.DbService {
	return ginsrv.MustGetService[*drugo.Drugo, *dbsvc.DbService](
		ctx,
		dbsvc.Name,
	)
}

// MustDefaultDB 返回默认数据库连接（default）。
//
// 该方法通常用于：
//   - 业务中不区分数据库
//   - 或作为业务库 / 公共库的基准连接
//
// 注意：
//   - 使用 Must 语义，内部若获取失败会直接 panic
//   - 适合在 HTTP 请求生命周期内使用
func MustDefaultDB(ginCtx *gin.Context) *gorm.DB {
	// 从 gin.Context 中获取标准 context.Context
	ctx := ginCtx.Request.Context()

	// 从 gin 上下文中获取数据库服务（若不存在会 panic）
	dbSvc := MustDB(ginCtx)

	// 获取默认数据库分组（default）
	pubGroup := dbSvc.Manager().MustGroup(consts.DbDefault)

	// 从默认分组中获取默认数据库连接
	db := pubGroup.MustGet(ctx, consts.DbDefault)

	return db
}

// MustPublicDB 返回公共库（public）中的默认数据库连接。
//
// 典型使用场景：
//   - 公共配置
//   - 公共字典
//   - 多业务共享的数据表
//
// 注意：
//   - 使用 Must 语义，任何错误都会 panic
//   - public 分组下默认数据库名仍为 default
func MustPublicDB(ginCtx *gin.Context) *gorm.DB {
	// 获取请求上下文
	ctx := ginCtx.Request.Context()

	// 获取数据库服务
	dbSvc := MustDB(ginCtx)

	// 获取公共数据库分组（public）
	pubGroup := dbSvc.Manager().MustGroup(consts.DbPublic)

	// 获取公共库中的默认数据库连接
	db := pubGroup.MustGet(ctx, consts.DbDefault)

	return db
}

// MustBusinessDB 返回指定业务数据库连接。
//
// 参数说明：
//   - ctx：Gin 请求上下文
//   - dbName：业务数据库名称（如 data_1、data_2 等）
//
// 实现说明：
//   - 业务库分组（business）下存在多个实际数据库
//   - 通过 mgorm.RegisterToDB 将 default 映射注册为指定业务库
//   - 后续即可通过逻辑名称直接获取对应数据库连接
//
// 注意：
//   - 使用 Must 语义，注册或获取失败会 panic
//   - 适用于分库、多租户业务场景
func MustBusinessDB(ctx *gin.Context, dbName string) *gorm.DB {
	// 逻辑数据库名（对外使用的名称）
	toName := dbName

	// 实际物理数据库名
	toDBName := dbName

	// 获取数据库服务
	dbSvc := MustDB(ctx)

	// 获取业务数据库分组（business）
	group := dbSvc.Manager().MustGroup(consts.DbBusiness)

	// 将默认数据库注册并映射为指定业务数据库
	mgorm.MustRegisterToDB(
		ctx.Request.Context(),
		group,
		consts.DbDefault, // 源数据库（模板）
		toName,           // 目标逻辑名称
		toDBName,         // 目标实际数据库名称
	)

	// 返回业务数据库连接
	return group.MustGet(ctx.Request.Context(), toName)
}

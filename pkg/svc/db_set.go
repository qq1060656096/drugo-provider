package svc

import "gorm.io/gorm"

// DBSet 表示一次业务执行过程中可用的一组数据库连接集合。
//
// 该结构用于在 drugo-provider 中统一传递数据库实例，
// 同时兼容 SaaS 与非 SaaS 两种场景，并支持普通 DB 与事务 DB（tx）。
//
// 设计为“集合（Set）”而非“上下文（Context）”，
// 表示它仅承载数据库资源本身，不包含请求、用户、租户等上下文信息，
// 更符合 Go 标准库中对资源集合的命名与职责划分。
type DBSet struct {
	// Default 默认数据库。
	//
	// 在非 SaaS 场景下，通常作为唯一的业务数据库使用；
	// 在 SaaS 场景下，可能用于框架级、基础能力或兜底访问。
	// 该字段既可以是普通 *gorm.DB，也可以是开启事务后的 tx。
	Default *gorm.DB

	// Public 公共数据库。
	//
	// 用于存放所有租户共享的数据，如基础配置、字典表、公共资源等。
	// 在非 SaaS 场景下，通常与 Default 指向同一个数据库实例。
	Public *gorm.DB

	// Business 业务数据库（租户数据库）。
	//
	// 在 SaaS 场景下，指向当前租户对应的业务库（如 data_1、data_2 等）；
	// 在非 SaaS 场景下，通常与 Default / Public 指向同一个数据库实例。
	// 同样支持普通 DB 或事务 DB（tx）。
	Business *gorm.DB
}

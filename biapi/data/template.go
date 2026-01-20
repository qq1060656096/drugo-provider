// Package data 提供 BI 模板的数据访问层。
// 本包仅负责数据库 CRUD 操作，不包含任何业务逻辑。
package data

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type baseData struct {
	Platform  int
	CompanyID int64
}

// Template 对应 bi_template 表的实体。
type Template struct {
	TemplateId int64      `gorm:"column:template_id;primaryKey;autoIncrement"`
	PlatformId int64      `gorm:"column:platform_id;not null"`
	CompanyID  int64      `gorm:"column:company_id;not null"`
	Code       string     `gorm:"column:code;type:varchar(64);not null"`
	Name       string     `gorm:"column:name;type:varchar(128)"`
	Status     int8       `gorm:"column:status;not null;default:1"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
}

// TableName 返回表名。
func (Template) TableName() string {
	return "bi_template"
}

// TemplateData 对应 bi_template_data 表的实体。
type TemplateData struct {
	TdId       int64      `gorm:"column:td_id;primaryKey;autoIncrement"`
	PlatformId int64      `gorm:"column:platform_id;not null"`
	TemplateId int64      `gorm:"column:template_id;not null"`
	CompanyId  int64      `gorm:"column:company_id;not null"`
	Env        string     `gorm:"column:env;type:enum('test','gray','prod');"`
	OpType     int        `gorm:"column:op_type;"`
	Content    string     `gorm:"column:content;type:mediumtext;not null"`
	Checksum   string     `gorm:"column:checksum;type:char(32);not null"`
	Status     int8       `gorm:"column:status;not null;default:1"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
}

// TableName 返回表名。
func (TemplateData) TableName() string {
	return "bi_template_data"
}

// templateRepo 是 TemplateRepo 的 GORM 实现。
type templateRepo struct {
}

// NewTemplateRepo 创建 TemplateRepo 实例。
func newTemplateRepo() *templateRepo {
	return &templateRepo{}
}

// FindByPlatformAndCode 根据平台 ID 和业务编码查询模板。
func (r *templateRepo) FindTpl(ctx context.Context, tplDb *gorm.DB, platId int64, code string) (*Template, error) {
	var tpl Template
	err := tplDb.WithContext(ctx).
		Where("platform_id = ?", platId).
		Where("code = ?", code).
		Where("status = 1").
		Where("deleted_at IS NULL").
		First(&tpl).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// FindTplData 根据模板 ID、环境和操作类型查询模板数据。
func (r *templateRepo) FindTplData(ctx context.Context, tplDb *gorm.DB, platId, tplId, cid int64, env string) (*TemplateData, error) {
	var data TemplateData
	err := tplDb.WithContext(ctx).
		Where("platform_id = ?", platId).
		Where("company_id in(0, ?)", cid).
		Where("template_id = ?", tplId).
		Where("env = ?", env).
		Where("status = 1").
		Where("deleted_at IS NULL").
		Order("company_id DESC").
		First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

package actor

import "github.com/jinzhu/gorm"

type Actor struct {
	CompanyID           uint32
	CompanyDatabaseName string
	AccountsID          uint32
	PublicDb            *gorm.DB
	BusinessDb          *gorm.DB
	TraceId             string
}

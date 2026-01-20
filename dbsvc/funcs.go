package dbsvc

import (
	"errors"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// ErrUnknownDriverType 当指定了不支持的数据库驱动类型时返回此错误。
var ErrUnknownDriverType = errors.New("mgormsvc: unknown driver type")

func CreateDialector(driverType, dsn string) (gorm.Dialector, error) {
	switch driverType {
	case "mysql":
		return mysql.Open(dsn), nil
	case "postgres":
		return postgres.Open(dsn), nil
	case "sqlite":
		return sqlite.Open(dsn), nil
	case "sqlserver":
		return sqlserver.Open(dsn), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownDriverType, driverType)
	}
}

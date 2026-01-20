package api

import "errors"

// 预定义的 API 错误。
var (
	ErrInvalidPlatformID = errors.New("api: invalid platform_id")
	ErrEmptyCode         = errors.New("api: code is required")
	ErrEmptyOpType       = errors.New("api: op_type is required")
	ErrInvalidOpType     = errors.New("api: invalid op_type, must be one of: list, detail, count, batchAdd, batchUpdate")
	ErrInvalidEnv        = errors.New("api: invalid env, must be one of: test, gray, prod")
)

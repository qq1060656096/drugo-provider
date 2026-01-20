package ginsrv

import (
	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/drugo/drugo"
	"github.com/qq1060656096/drugo/kernel"
)

// AppVarFromContext 从 gin context 中获取指定类型的应用变量。
// 如果 context 中未设置应用变量，返回 ErrAppNotFound；
// 如果类型断言失败，返回 ErrAppTypeMismatch。
func AppVarFromContext[T any](ctx *gin.Context) (T, error) {
	var zero T
	app, ok := ctx.Get(drugo.Name)
	if !ok {
		return zero, ErrAppNotFound
	}
	appVar, ok := app.(T)
	if !ok {
		return zero, ErrAppTypeMismatch
	}
	return appVar, nil
}

// MustAppVarFromContext 与 AppVarFromContext 功能相同，但在发生错误时会 panic。
func MustAppVarFromContext[T kernel.Kernel](ctx *gin.Context) T {
	appVar, err := AppVarFromContext[T](ctx)
	if err != nil {
		panic(err)
	}
	return appVar
}

// GetService 从 gin context 中获取指定类型的服务。
// T 为应用内核类型，S 为需要获取的服务类型。
// 如果获取应用变量失败或服务不存在，返回对应的错误。
func GetService[T kernel.Kernel, S kernel.Service](ctx *gin.Context, name string) (S, error) {
	var zero S
	appVar, err := AppVarFromContext[T](ctx)
	if err != nil {
		return zero, err
	}

	svc, err := kernel.GetService[S](appVar, name)
	if err != nil {
		return zero, err
	}
	return svc, nil
}

// MustGetService 与 GetService 功能相同，但在发生错误时会 panic。
func MustGetService[T kernel.Kernel, S kernel.Service](ctx *gin.Context, name string) S {
	svc, err := GetService[T, S](ctx, name)
	if err != nil {
		panic(err)
	}
	return svc
}

package svc

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/qq1060656096/drugo-provider/ginsrv"
	"github.com/qq1060656096/drugo-provider/i18nsvc"
	"github.com/qq1060656096/drugo/drugo"
)

// MustI18n 从 Gin 上下文中获取国际化服务（I18nService）。
//
// 该方法是对 ginsrv.MustGetService 的语义化封装，
// 用于在 HTTP 请求生命周期内获取当前应用绑定的国际化服务实例。
//
// 设计目的：
//   - 屏蔽底层 Service 获取细节，提升业务代码可读性
//   - 统一 I18nService 的获取入口，避免在业务层直接依赖 ginsrv
//
// 使用约定：
//   - ctx 必须是当前请求对应的 *gin.Context
//   - I18nService 需已在 Gin 中间件阶段完成注册
//
// 注意事项：
//   - 使用 Must 语义：
//   - 当服务未注册
//   - 或类型断言失败
//   - 或上下文不合法
//     时将直接 panic
//   - 适合在 Handler / Service / Data 层中使用，
//     不建议在可恢复错误路径中调用
func MustI18n(ctx *gin.Context) *i18nsvc.I18nService {
	return ginsrv.MustGetService[*drugo.Drugo, *i18nsvc.I18nService](
		ctx,
		i18nsvc.Name,
	)
}

// MustT 根据指定的语言和键获取翻译文本。
//
// 该方法是对 i18nsvc.I18nService.T 的语义化封装，
// 提供更便捷的翻译调用方式。
//
// 参数说明：
//   - ctx：Gin 请求上下文
//   - lang：目标语言代码，如果为空则使用默认语言
//   - key：翻译键名
//   - data：模板数据，用于替换翻译文本中的占位符
//
// 返回：翻译后的文本，如果翻译失败则返回键名
//
// 使用示例：
//
//	message := svc.MustT(c, "zh", "welcome", nil)
//	data := map[string]any{"Name": "张三"}
//	greeting := svc.MustT(c, "zh", "greeting", data)
func MustT(ctx *gin.Context, lang, key string, data map[string]any) string {
	i18nSvc := MustI18n(ctx)
	return i18nSvc.T(lang, key, data)
}

// MustTCtx 从context中获取语言信息并翻译文本。
//
// 该方法是对 i18nsvc.I18nService.TCtx 的语义化封装，
// 适用于已经设置了语言信息的context。
//
// 参数说明：
//   - ctx：Gin 请求上下文（需包含语言信息）
//   - key：要翻译的文本键
//   - data：模板变量
//
// 返回：翻译后的文本
//
// 使用示例：
//
//	// 首先设置语言信息
//	ctx = svc.MustI18n(c).WithLang(c.Request.Context(), "zh")
//	// 然后使用TCtx进行翻译
//	message := svc.MustTCtx(c, "welcome", nil)
func MustTCtx(ctx *gin.Context, key string, data map[string]any) string {
	i18nSvc := MustI18n(ctx)
	return i18nSvc.TCtx(ctx.Request.Context(), key, data)
}

// MustWithLang 将语言信息写入context。
//
// 该方法是对 i18nsvc.I18nService.WithLang 的语义化封装，
// 用于在请求处理过程中设置语言偏好。
//
// 参数说明：
//   - ctx：Gin 请求上下文
//   - lang：语言代码（如 "zh", "en" 等）
//
// 返回：包含语言信息的 context.Context
//
// 使用示例：
//
//	ctxWithLang := svc.MustWithLang(c, "zh")
//	c.Request = c.Request.WithContext(ctxWithLang)
func MustWithLang(ctx *gin.Context, lang string) context.Context {
	i18nSvc := MustI18n(ctx)
	return i18nSvc.WithLang(ctx.Request.Context(), lang)
}

// MustLang 从context中获取语言信息。
//
// 该方法是对 i18nsvc.I18nService.Lang 的语义化封装，
// 用于获取当前请求的语言设置。
//
// 参数说明：
//   - ctx：Gin 请求上下文
//
// 返回：语言代码字符串
//
// 使用示例：
//
//	lang := svc.MustLang(c)
//	if lang == "zh" {
//	    // 中文处理逻辑
//	}
func MustLang(ctx *gin.Context) string {
	i18nSvc := MustI18n(ctx)
	return i18nSvc.Lang(ctx.Request.Context())
}

// MustGetSupportedLanguages 获取支持的语言列表。
//
// 该方法是对 i18nsvc.I18nService.GetSupportedLanguages 的语义化封装，
// 用于获取当前系统中支持的所有语言。
//
// 参数说明：
//   - ctx：Gin 请求上下文
//
// 返回：支持的语言代码列表
//
// 使用示例：
//
//	languages := svc.MustGetSupportedLanguages(c)
//	fmt.Printf("支持的语言: %v", languages)
func MustGetSupportedLanguages(ctx *gin.Context) []string {
	i18nSvc := MustI18n(ctx)
	return i18nSvc.GetSupportedLanguages()
}

// MustReloadI18n 重新加载翻译文件。
//
// 该方法是对 i18nsvc.I18nService.Reload 的语义化封装，
// 用于在翻译文件更新后重新加载。
//
// 参数说明：
//   - ctx：Gin 请求上下文
//
// 返回：错误信息（如果重新加载失败）
//
// 使用示例：
//
//	// 在翻译文件更新后调用
//	if err := svc.MustReloadI18n(c); err != nil {
//	    log.Printf("重新加载翻译文件失败: %v", err)
//	}
func MustReloadI18n(ctx *gin.Context) error {
	i18nSvc := MustI18n(ctx)
	return i18nSvc.Reload()
}

# i18nsvc

基于 `mi18n` 的国际化服务，提供多语言翻译支持。

## 特性

- 🌍 支持多种语言翻译
- 📄 支持多种文件格式：JSON、YAML、TOML
- 🔒 并发安全，使用读写锁保护
- 🎯 支持模板变量替换
- 📦 支持从Context中获取语言信息
- 🚀 轻量级，零依赖外部运行时

## 安装

```bash
go get github.com/qq1060656096/drugo-provider/i18nsvc
```

## 配置

```yaml
i18n:
  locale_dir: "locales"          # 翻译文件目录
  default_lang: "en"             # 默认语言
```

## 翻译文件格式

### JSON 格式

```json
[
  { "id": "welcome", "translation": "欢迎" },
  { "id": "greeting", "translation": "你好，{{.Name}}！" }
]
```

### YAML 格式

```yaml
- id: welcome
  translation: 欢迎
- id: greeting
  translation: 你好，{{.Name}}！
```

### TOML 格式

```toml
[[translations]]
id = "welcome"
translation = "欢迎"

[[translations]]
id = "greeting"
translation = "你好，{{.Name}}！"
```

## 使用示例

### 服务注册

```go
import (
    "github.com/qq1060656096/drugo/drugo"
    "github.com/qq1060656096/drugo-provider/i18nsvc"
)

func main() {
    app := drugo.New()
    
    // 注册i18n服务
    app.Register(i18nsvc.New())
    
    // 启动应用
    if err := app.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### 基本使用

```go
import (
    "github.com/qq1060656096/drugo-provider/ginsrv"
    "github.com/qq1060656096/drugo-provider/i18nsvc"
)

// 获取i18n服务
i18nSvc := ginsrv.MustGetService[*drugo.Drugo, *i18nsvc.I18nService](c, i18nsvc.Name)

// 基本翻译
welcome := i18nSvc.T("zh", "welcome", nil) // 输出: 欢迎

// 带变量的翻译
data := map[string]any{"Name": "张三"}
greeting := i18nSvc.T("zh", "greeting", data) // 输出: 你好，张三！

// 使用Context进行翻译
ctx := i18nSvc.WithLang(c.Request.Context(), "zh")
welcome := i18nSvc.TCtx(ctx, "welcome", nil) // 输出: 欢迎

// 从Context获取语言
lang := i18nSvc.Lang(ctx) // 输出: zh
```

### 在Gin中间件中使用

```go
// 设置语言中间件
func SetLanguageMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从请求头、Cookie或参数中获取语言
        lang := c.GetHeader("Accept-Language")
        if lang == "" {
            lang = "en" // 默认语言
        }
        
        // 获取i18n服务
        i18nSvc := ginsrv.MustGetService[*drugo.Drugo, *i18nsvc.I18nService](c, i18nsvc.Name)
        
        // 将语言信息写入Context
        c.Request = c.Request.WithContext(i18nSvc.WithLang(c.Request.Context(), lang))
        c.Next()
    }
}

// 在处理器中使用
func handler(c *gin.Context) {
    i18nSvc := ginsrv.MustGetService[*drugo.Drugo, *i18nsvc.I18nService](c, i18nsvc.Name)
    
    message := i18nSvc.TCtx(c.Request.Context(), "welcome", nil)
    c.JSON(200, gin.H{"message": message})
}
```

## API 文档

### I18nService 方法

#### T(lang, key string, data map[string]any) string

根据指定的语言和键获取翻译文本。

- `lang`: 目标语言代码，如果为空则使用默认语言
- `key`: 翻译键名
- `data`: 模板数据，用于替换翻译文本中的占位符
- 返回: 翻译后的文本，如果翻译失败则返回键名

#### TCtx(ctx context.Context, key string, data map[string]any) string

从context中获取语言信息并翻译文本。

- `ctx`: 包含语言信息的context
- `key`: 要翻译的文本键
- `data`: 模板变量
- 返回: 翻译后的文本

#### WithLang(ctx context.Context, lang string) context.Context

将语言信息写入context。

#### Lang(ctx context.Context) string

从context中获取语言信息。

#### I18n() *mi18n.I18n

返回底层的 mi18n.I18n 实例。

#### GetSupportedLanguages() []string

返回支持的语言列表。该方法会扫描locale目录下的所有翻译文件，返回支持的语言代码。

#### Reload() error

重新加载翻译文件。当翻译文件更新后，可以调用此方法重新加载。

## 高级用法

### 获取支持的语言

```go
// 获取支持的语言列表
languages := i18nSvc.GetSupportedLanguages()
fmt.Printf("支持的语言: %v", languages) // 输出: [zh en ja]
```

### 热重载翻译文件

```go
// 当翻译文件更新后，可以重新加载
if err := i18nSvc.Reload(); err != nil {
    log.Printf("重新加载翻译文件失败: %v", err)
}
```

### 在便捷函数中使用新功能

```go
// 获取支持的语言
languages := svc.MustGetSupportedLanguages(c)

// 重新加载翻译文件
if err := svc.MustReloadI18n(c); err != nil {
    log.Printf("重新加载失败: %v", err)
}
```

## 语言代码规范

建议使用标准的语言代码：

- `zh`: 中文
- `en`: 英文
- `ja`: 日文
- `ko`: 韩文
- `fr`: 法文
- `de`: 德文
- `es`: 西班牙文
- `ru`: 俄文

## 模板变量

翻译文本支持Go模板语法，可以使用变量替换：

```json
[
  { "id": "user_info", "translation": "用户：{{.Name}}，年龄：{{.Age}}" }
]
```

使用时：

```go
data := map[string]any{
    "Name": "张三",
    "Age": 25,
}
result := i18nSvc.T("zh", "user_info", data)
// 输出: 用户：张三，年龄：25
```

## 并发安全

i18nsvc 是并发安全的，可以在多个goroutine中同时使用。

## 测试

运行测试：

```bash
go test ./i18nsvc
```

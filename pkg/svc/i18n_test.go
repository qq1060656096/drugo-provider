package svc

import (
	"testing"
)

func TestMustI18n(t *testing.T) {
	// 这个测试需要完整的kernel环境，这里只做基本的结构测试
	// 实际测试需要集成测试环境
	t.Skip("需要完整的kernel环境进行集成测试")
}

func TestMustT(t *testing.T) {
	t.Skip("需要完整的kernel环境进行集成测试")
}

func TestMustTCtx(t *testing.T) {
	t.Skip("需要完整的kernel环境进行集成测试")
}

func TestMustWithLang(t *testing.T) {
	t.Skip("需要完整的kernel环境进行集成测试")
}

func TestMustLang(t *testing.T) {
	t.Skip("需要完整的kernel环境进行集成测试")
}

func TestMustGetSupportedLanguages(t *testing.T) {
	t.Skip("需要完整的kernel环境进行集成测试")
}

func TestMustReloadI18n(t *testing.T) {
	t.Skip("需要完整的kernel环境进行集成测试")
}

// 以下为集成测试示例，需要完整的drugo环境
func TestI18nIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 这里可以编写完整的集成测试
	// 需要启动完整的drugo应用
	t.Skip("集成测试需要完整的应用环境")
}

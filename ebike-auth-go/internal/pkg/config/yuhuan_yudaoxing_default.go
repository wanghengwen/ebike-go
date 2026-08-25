//go:build !yuhuan_yudaoxing

package config

// YuhuanYudaoxingCompileDefault 玉环公共电单车客户定制的编译期默认值。
// 使用 -tags=yuhuan_yudaoxing 编译时，默认开启玉岛行登录定制逻辑。
func YuhuanYudaoxingCompileDefault() bool {
	return false
}

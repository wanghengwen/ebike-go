//go:build yuhuan_yudaoxing

package config

// YuhuanYudaoxingCompileDefault 玉环公共电单车客户定制的编译期默认值。
// 本文件在 go build -tags=yuhuan_yudaoxing 时生效，默认开启定制逻辑。
func YuhuanYudaoxingCompileDefault() bool {
	return true
}

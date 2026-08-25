package config

// nacosYudaoxingProperties 解析 Nacos ebike-auth-client.yml 中的 yudaoxing 段。
// Java Spring Boot 使用 kebab-case（base-url、merchant-private-key 等），
// Go 同时兼容 kebab-case 与 camelCase，kebab-case 优先。
type nacosYudaoxingProperties struct {
	Account string `yaml:"account"`

	BaseUrlKebab string `yaml:"base-url"`
	BaseUrlCamel string `yaml:"baseUrl"`

	MerchantPrivateKebab string `yaml:"merchant-private-key"`
	MerchantPrivateCamel string `yaml:"merchantPrivateKey"`

	ServerPublicKebab string `yaml:"server-public-key"`
	ServerPublicCamel string `yaml:"serverPublicKey"`
}

func (p nacosYudaoxingProperties) toYudaoxingConfig() YudaoxingConfig {
	return YudaoxingConfig{
		Account:            p.Account,
		BaseUrl:            firstNonEmpty(p.BaseUrlKebab, p.BaseUrlCamel),
		MerchantPrivateKey: firstNonEmpty(p.MerchantPrivateKebab, p.MerchantPrivateCamel),
		ServerPublicKey:    firstNonEmpty(p.ServerPublicKebab, p.ServerPublicCamel),
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

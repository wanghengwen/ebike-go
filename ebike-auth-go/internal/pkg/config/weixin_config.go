package config

// 微信公众平台默认凭据，与 Java test 分支 JsApiSignatureService 中
// @Value("${weixin.publicplatform.appid:...}") 的硬编码默认值保持一致。
// 生产环境若玉环使用独立公众号，务必在 Nacos ebike-auth-client.yml 的
// weixin.publicplatform 下覆盖，否则 wx.config 签名会与前端公众号不匹配。
const (
	defaultWeixinPublicAppID     = "wx88578acc4c9e4046"
	defaultWeixinPublicAppSecret = "1debf29f45bfcf547072570b719bfa3c"
)

// nacosWeixinProperties 解析 Nacos ebike-auth-client.yml 中的微信公众平台配置。
// 对应 Java test 分支 JsApiSignatureService（feature/20250614-jsapi签名添加）。
type nacosWeixinProperties struct {
	PublicPlatform struct {
		AppIdKebab     string `yaml:"app-id"`
		AppIdCamel     string `yaml:"appid"`
		AppSecretKebab string `yaml:"app-secret"`
		AppSecretCamel string `yaml:"appSecret"`
	} `yaml:"publicplatform"`
}

type WeixinPublicPlatformConfig struct {
	AppId      string
	AppSecret  string
	APIBaseURL string
}

func (p nacosWeixinProperties) toWeixinPublicPlatformConfig() WeixinPublicPlatformConfig {
	return WeixinPublicPlatformConfig{
		AppId:     firstNonEmpty(p.PublicPlatform.AppIdKebab, p.PublicPlatform.AppIdCamel),
		AppSecret: firstNonEmpty(p.PublicPlatform.AppSecretKebab, p.PublicPlatform.AppSecretCamel),
	}
}

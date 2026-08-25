package model

// TenantConfig maps to t_tenant_config
type TenantConfig struct {
	ID         int64  `gorm:"column:id;primaryKey"`
	TenantId   string `gorm:"column:tenant_id"`
	SupplierId int64  `gorm:"column:supplier_id"`
	Type       int    `gorm:"column:type"`        // 1=短信, 2=语音
	PayPattern int    `gorm:"column:pay_pattern"` // 付费模式
	IzDel      int    `gorm:"column:iz_del"`
}

func (TenantConfig) TableName() string { return "t_tenant_config" }

// Supplier maps to t_supplier
type Supplier struct {
	ID         int64  `gorm:"column:id;primaryKey"`
	Channel    int    `gorm:"column:channel"`     // 1=阿里云, 2=创蓝
	Type       int    `gorm:"column:type"`        // 1=短信, 2=语音
	ConfigJson string `gorm:"column:config_json"` // JSON格式的供应商配置
	IzDel      int    `gorm:"column:iz_del"`
}

func (Supplier) TableName() string { return "t_supplier" }

// Template maps to t_template
type Template struct {
	ID                   int64  `gorm:"column:id;primaryKey"`
	SupplierId           int64  `gorm:"column:supplier_id"`
	Type                 int    `gorm:"column:type"`
	TemplateName         string `gorm:"column:template_name"`
	TemplateContent      string `gorm:"column:template_content"`
	TemplateCode         string `gorm:"column:template_code"`
	SupplierTemplateCode string `gorm:"column:supplier_template_code"` // 供应商侧模板编码
	IzDel                int    `gorm:"column:iz_del"`
}

func (Template) TableName() string { return "t_template" }

// MessageSign maps to t_message_sign
type MessageSign struct {
	ID       int64  `gorm:"column:id;primaryKey"`
	TenantId string `gorm:"column:tenant_id"`
	Sign     string `gorm:"column:sign"`
	IzDel    int    `gorm:"column:iz_del"`
}

func (MessageSign) TableName() string { return "t_message_sign" }

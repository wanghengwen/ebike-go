package model

// SendMessageReq is the request body for sending an SMS message.
type SendMessageReq struct {
	TraceId      string            `json:"traceId" binding:"required"`
	Source       string            `json:"source" binding:"required"`
	TenantId     string            `json:"tenantId" binding:"required"`
	Phone        string            `json:"phone" binding:"required,china_phone"`
	TemplateCode string            `json:"templateCode" binding:"required"`
	SignName     string            `json:"signName" binding:"required"`
	Params       map[string]string `json:"params"`
	Secret       string            `json:"secret" binding:"required"`
}

// SendVoiceReq is the request body for sending a voice notification.
type SendVoiceReq struct {
	TraceId      string            `json:"traceId" binding:"required"`
	Source       string            `json:"source" binding:"required"`
	TenantId     string            `json:"tenantId" binding:"required"`
	Phone        string            `json:"phone" binding:"required,china_phone"`
	TemplateCode string            `json:"templateCode" binding:"required"`
	Params       map[string]string `json:"params"`
	Secret       string            `json:"secret" binding:"required"`
}

// SendResult represents the outcome of a send operation.
type SendResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

// SuccessResult creates a successful SendResult with the given ID.
func SuccessResult(id string) *SendResult {
	return &SendResult{
		Success: true,
		Message: "成功",
		ID:      id,
	}
}

// FailResult creates a failed SendResult with the given message.
func FailResult(msg string) *SendResult {
	return &SendResult{
		Success: false,
		Message: msg,
	}
}

// FailResultWithID creates a failed SendResult with both a message and an ID.
func FailResultWithID(msg string, id string) *SendResult {
	return &SendResult{
		Success: false,
		Message: msg,
		ID:      id,
	}
}

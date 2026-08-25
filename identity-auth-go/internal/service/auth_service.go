package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"identity-auth-go/internal/model"
	"identity-auth-go/internal/pkg/cache"
	"identity-auth-go/internal/pkg/mask"
	ossutil "identity-auth-go/internal/pkg/oss"
	goredis "identity-auth-go/internal/pkg/redis"
	"identity-auth-go/internal/tps"

	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

// --------------------------------------------------------------------------
// Shadow / dry-run helpers
// --------------------------------------------------------------------------

// RecordAuth saves auth request to shadow record list (dry-run mode).
func (s *AuthService) RecordAuth(path string, req model.AuthRequest) error {
	payload, _ := json.Marshal(req)
	record := &model.ShadowRecordList{
		RequestPath: path,
		Payload:     string(payload),
		TenantId:    req.TenantId,
		TraceId:     req.TraceId,
		Type:        req.Type,
		CreatedAt:   time.Now(),
	}
	return model.InsertShadowRecord(s.db, record)
}

// RecordCharge saves charge request to shadow record list (dry-run mode).
func (s *AuthService) RecordCharge(path string, req model.ChargeRequest) error {
	payload, _ := json.Marshal(req)
	record := &model.ShadowRecordList{
		RequestPath: path,
		Payload:     string(payload),
		TenantId:    req.TenantId,
		TraceId:     req.TraceId,
		Type:        req.Type,
		CreatedAt:   time.Now(),
	}
	return model.InsertShadowRecord(s.db, record)
}

// --------------------------------------------------------------------------
// Production auth flow (matching Java AuthServiceImpl)
// --------------------------------------------------------------------------

// Auth performs the full authentication flow.
func (s *AuthService) Auth(req model.AuthRequest) *model.R {
	tc, err := cache.GetTenantConfig(s.db, req.TenantId, req.Type)
	if err != nil || tc == nil {
		msg := "租户配置不存在"
		if err != nil {
			msg = err.Error()
		}
		log.Printf("[AuthService] traceId=%s, tenant config not found: tenantId=%s, type=%d, msg=%s", req.TraceId, req.TenantId, req.Type, msg)
		return &model.R{Success: false, Msg: msg}
	}

	supplierId := tc.SupplierId
	pattern := tc.Pattern

	// Check prepaid balance
	if isPrepaid(pattern) {
		hasCallTimes, err := goredis.GetCallTimes(req.TenantId, req.Type)
		if err != nil {
			log.Printf("[AuthService] traceId=%s, redis error checking call times: %v", req.TraceId, err)
		}
		if hasCallTimes <= 0 {
			log.Printf("[AuthService] traceId=%s, tenantId=%s, prepaid call times exhausted", req.TraceId, req.TenantId)
			// Java's quirk: type=1 returns failure, type=2 returns success
			if req.Type == tps.AuthTypeIdentityCard {
				return &model.R{Success: false, Msg: "预付费第三方实名认证api调用次数不足"}
			}
			// type=2: return success with izSame=true, score=100
			return &model.R{Success: true, Data: tps.AuthResponse{IzSame: true, Score: "100"}}
		}
	}

	return s.doAuth(req, supplierId)
}

func (s *AuthService) doAuth(req model.AuthRequest, supplierId int64) *model.R {
	switch req.Type {
	case tps.AuthTypeIdentityCard:
		return s.matchIdCard(req, supplierId)
	case tps.AuthTypeIdentityCardFace:
		return s.matchIdCardFace(req, supplierId)
	default:
		return &model.R{Success: false, Msg: "不支持的认证类型"}
	}
}

// matchIdCard performs two-element authentication (name + ID card).
func (s *AuthService) matchIdCard(req model.AuthRequest, supplierId int64) *model.R {
	// Check Redis cache first
	cachedResult, exists, _ := goredis.GetAuthResult(req.IdCardNum, req.Name)
	if exists {
		logAuthCacheHit(req, tps.AuthTypeIdentityCard, cachedResult, "")
		if cachedResult {
			return &model.R{Success: true}
		}
		return &model.R{Success: false, Msg: "身份校验未通过"}
	}

	// Get supplier
	supplier, err := cache.GetSupplier(s.db, supplierId)
	if err != nil || supplier == nil {
		msg := fmt.Sprintf("提供商[%d]不存在", supplierId)
		if err != nil {
			msg = err.Error()
		}
		log.Printf("[AuthService] traceId=%s, supplier not found: %d, msg=%s", req.TraceId, supplierId, msg)
		return &model.R{Success: false, Msg: msg}
	}

	// Call TPS
	tpsAuth := tps.GetTpsIdAuth(supplier.Channel)
	if tpsAuth == nil {
		return &model.R{Success: false, Msg: "供应商处理bean不存在"}
	}

	authParam := tps.CommonTpsIdAuthParam{
		TraceId:   req.TraceId,
		AuthType:  req.Type,
		Name:      req.Name,
		IdCardNum: req.IdCardNum,
		Image:     req.Image,
		Config:    supplier.Config,
	}
	responseData := make(map[string]interface{})
	authResult := tpsAuth.Authenticate(authParam, responseData)
	logAuthTPSResult(req, supplierId, req.Type, authResult, responseData)

	// Decrease call times (async)
	go s.decreaseCallTimes(req)

	// Get authResponse
	authResp := extractAuthResponse(responseData)

	// Save call record
	go s.saveCallRecord(req, supplierId, responseData, req.Type)

	// Cache result in Redis
	_ = goredis.SetAuthResult(req.IdCardNum, req.Name, authResult)

	if authResult {
		return &model.R{Success: true}
	}
	msg := "身份校验未通过"
	if authResp != nil && authResp.Info != "" {
		msg = authResp.Info
	}
	return &model.R{Success: false, Msg: msg}
}

// matchIdCardFace performs three-element authentication (name + ID card + face).
func (s *AuthService) matchIdCardFace(req model.AuthRequest, supplierId int64) *model.R {
	// Step 1: Try local face match with cached matched face image
	result := s.matchFace(req, true)
	if result != nil {
		return result
	}

	// Step 2: Try TPS ID card + face auth
	return s.doMatchIdCardFace(req, supplierId)
}

// matchFace tries to match against a previously stored face image.
func (s *AuthService) matchFace(req model.AuthRequest, match bool) *model.R {
	url := ossutil.GetUrl(match, req.IdCardNum, req.Name)
	cacheImg, ok := ossutil.UrlToBase64(url)
	if !ok {
		return nil // No cached image
	}

	r := s.doMatchFaceImage(req, cacheImg)
	if r == nil {
		return nil
	}

	if match {
		return r
	}

	// For no_match path: if face matches the "no_match" image, it means failure
	if r.Success {
		if data, ok := r.Data.(map[string]interface{}); ok {
			if izSame, _ := data["izSame"].(bool); izSame {
				return &model.R{Success: true, Data: tps.AuthResponse{IzSame: false, Score: "0", Info: "人脸校验未通过"}}
			}
		}
	}
	return nil
}

// doMatchFaceImage performs face comparison via TPS.
func (s *AuthService) doMatchFaceImage(req model.AuthRequest, targetImage string) *model.R {
	tc, err := cache.GetTenantConfig(s.db, req.TenantId, tps.AuthTypeFaceMatch)
	if err != nil || tc == nil {
		return nil
	}

	supplier, err := cache.GetSupplier(s.db, tc.SupplierId)
	if err != nil || supplier == nil {
		return nil
	}

	tpsAuth := tps.GetTpsIdAuth(supplier.Channel)
	if tpsAuth == nil {
		return nil
	}

	faceParam := tps.CommonTpsFaceMatchParam{
		Config:    supplier.Config,
		TraceId:   req.TraceId,
		Name:      req.Name,
		IdCardNum: req.IdCardNum,
		SourceImg: req.Image,
		TargetImg: targetImage,
	}
	responseData := make(map[string]interface{})
	authResult := tpsAuth.MatchFaceImage(faceParam, responseData)

	log.Printf("[AuthService] traceId=%s tenantId=%s type=3 supplierId=%d name=%s idCard=%s tps faceMatch status=%v",
		req.TraceId, req.TenantId, supplier.ID, req.Name, mask.MaskIDCard(req.IdCardNum), responseData["status"])

	imageUrl := ossutil.GetUrl(authResult, req.IdCardNum, req.Name)
	responseData["imageUrl"] = imageUrl

	authResp := extractAuthResponse(responseData)
	isException := fmt.Sprintf("%v", responseData["status"]) == "3"

	// Decrease call times for HuaWei channel
	if supplier.Channel == tps.ChannelHuaWei {
		go s.decreaseCallTimes(req)
	}

	// Save call record
	go s.saveCallRecord(req, supplier.ID, responseData, tps.AuthTypeFaceMatch)

	if isException {
		return nil
	}

	// Java: status==2 -> failure; any other non-3 status -> success.
	if fmt.Sprintf("%v", responseData["status"]) == "2" {
		msg := "人脸校验未通过"
		if authResp != nil && authResp.Info != "" {
			msg = authResp.Info
		}
		return &model.R{Success: false, Msg: msg}
	}

	if authResp != nil {
		return &model.R{Success: true, Data: authResp}
	}
	return &model.R{Success: true}
}

// doMatchIdCardFace performs TPS-based ID card + face verification.
func (s *AuthService) doMatchIdCardFace(req model.AuthRequest, supplierId int64) *model.R {
	// Check cached two-element result first.
	// Java: cached false -> failure; cached true -> Mono.empty() -> still calls TPS authenticate.
	cachedResult, exists, _ := goredis.GetAuthResult(req.IdCardNum, req.Name)
	if exists && !cachedResult {
		logAuthCacheHit(req, tps.AuthTypeIdentityCardFace, false, " cached二要素=false")
		return &model.R{Success: true, Data: tps.AuthResponse{IzSame: false, Score: "0", Info: "身份校验未通过"}}
	}

	// Get supplier
	supplier, err := cache.GetSupplier(s.db, supplierId)
	if err != nil || supplier == nil {
		msg := fmt.Sprintf("提供商[%d]不存在", supplierId)
		if err != nil {
			msg = err.Error()
		}
		return &model.R{Success: false, Msg: msg}
	}

	tpsAuth := tps.GetTpsIdAuth(supplier.Channel)
	if tpsAuth == nil {
		return &model.R{Success: false, Msg: "供应商处理bean不存在"}
	}

	authParam := tps.CommonTpsIdAuthParam{
		TraceId:   req.TraceId,
		AuthType:  req.Type,
		Name:      req.Name,
		IdCardNum: req.IdCardNum,
		Image:     req.Image,
		Config:    supplier.Config,
	}
	responseData := make(map[string]interface{})
	authResult := tpsAuth.Authenticate(authParam, responseData)
	logAuthTPSResult(req, supplierId, req.Type, authResult, responseData)

	isException := fmt.Sprintf("%v", responseData["status"]) == "3"
	authResp := extractAuthResponse(responseData)

	// Decrease call times
	go s.decreaseCallTimes(req)

	// Upload face image to OSS
	go func() {
		url, err := ossutil.UploadFaceImg(authResult, req.IdCardNum, req.Name, req.Image)
		if err != nil {
			log.Printf("[AuthService] traceId=%s, upload face img error: %v", req.TraceId, err)
		} else {
			responseData["imageUrl"] = url
		}
		// Save call record after upload
		s.saveCallRecord(req, supplierId, responseData, req.Type)
	}()

	if isException {
		msg := "认证异常"
		if authResp != nil && authResp.Info != "" {
			msg = authResp.Info
		}
		return &model.R{Success: false, Msg: msg}
	}

	if authResp != nil {
		return &model.R{Success: true, Data: authResp}
	}
	return &model.R{Success: true}
}

// --------------------------------------------------------------------------
// Charge flow (matching Java ChargeServiceImpl)
// --------------------------------------------------------------------------

// Charge processes a charge request.
func (s *AuthService) Charge(req model.ChargeRequest) *model.R {
	tc, err := cache.GetTenantConfig(s.db, req.TenantId, req.Type)
	if err != nil || tc == nil {
		msg := "租户配置不存在"
		if err != nil {
			msg = err.Error()
		}
		log.Printf("[AuthService] traceId=%s, charge tenant config not found: tenantId=%s, type=%d, msg=%s", req.TraceId, req.TenantId, req.Type, msg)
		return &model.R{Success: false, Msg: msg}
	}

	// Postpaid cannot charge
	if tc.Pattern == 2 {
		return &model.R{Success: false, Msg: "后付费不能充值余量"}
	}

	// Java fire-and-forget: save record and increase Redis async, always return success.
	chargeRecord := model.ChargeRecord{
		TenantId: req.TenantId,
		Amount:   req.Amount,
		Quantity: req.Quantity,
		Type:     req.Type,
	}
	chargeRecord.TraceId = req.TraceId
	go func() {
		if err := s.db.Create(&chargeRecord).Error; err != nil {
			log.Printf("[AuthService] traceId=%s, save charge record error: %v", req.TraceId, err)
			return
		}
		log.Printf("[AuthService] traceId=%s, save charge record success", req.TraceId)
		newBalance, err := goredis.IncreaseCallTimes(req.TenantId, req.Type, req.Quantity)
		if err != nil {
			log.Printf("[AuthService] traceId=%s, increase call times error: %v", req.TraceId, err)
		} else {
			log.Printf("[AuthService] traceId=%s, tenantId=%s, charge done, newBalance=%d", req.TraceId, req.TenantId, newBalance)
		}
	}()

	return &model.R{Success: true}
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func (s *AuthService) decreaseCallTimes(req model.AuthRequest) {
	tc, err := cache.GetTenantConfig(s.db, req.TenantId, req.Type)
	if err != nil || tc == nil {
		return
	}
	if tc.Pattern != 1 {
		return // Not prepaid
	}
	_, err = goredis.DecreaseCallTimes(req.TenantId, req.Type)
	if err != nil {
		log.Printf("[AuthService] traceId=%s, decrease call times error: %v", req.TraceId, err)
	}
}

func (s *AuthService) saveCallRecord(req model.AuthRequest, supplierId int64, responseData map[string]interface{}, authType int) {
	tableName := ""
	switch authType {
	case tps.AuthTypeIdentityCard:
		tableName = "t_two_call_record_" + time.Now().Format("2006_01")
	case tps.AuthTypeIdentityCardFace:
		tableName = "t_three_call_record_" + time.Now().Format("2006_01")
	case tps.AuthTypeFaceMatch:
		tableName = "t_face_match_record_" + time.Now().Format("2006_01")
	default:
		log.Printf("[AuthService] unsupported auth type for call record: %d", authType)
		return
	}

	uniqueId := ""
	if v, ok := responseData["uniqueId"].(string); ok {
		uniqueId = v
	}
	responseJSON := "{}"
	if v, ok := responseData["response"]; ok {
		b, _ := json.Marshal(v)
		responseJSON = string(b)
	}
	imageUrl := ""
	if v, ok := responseData["imageUrl"].(string); ok {
		imageUrl = v
	}
	status := 0
	if v, ok := responseData["status"]; ok {
		switch vt := v.(type) {
		case int:
			status = vt
		case float64:
			status = int(vt)
		}
	}
	score := parseScoreToInt(responseData["score"])

	callAt := time.Now().Format("2006-01-02 15:04:05")

	var sql string
	var args []interface{}

	switch authType {
	case tps.AuthTypeIdentityCard:
		sql = fmt.Sprintf("INSERT INTO %s(tenant_id,supplier_id,name,identity_no,call_at,status,unique_id,response_data,trace_id) VALUES(?,?,?,?,?,?,?,?,?)", tableName)
		args = []interface{}{req.TenantId, supplierId, req.Name, req.IdCardNum, callAt, status, uniqueId, responseJSON, req.TraceId}
	case tps.AuthTypeIdentityCardFace:
		sql = fmt.Sprintf("INSERT INTO %s(tenant_id,supplier_id,name,identity_no,image_url,call_at,status,unique_id,response_data,trace_id) VALUES(?,?,?,?,?,?,?,?,?,?)", tableName)
		args = []interface{}{req.TenantId, supplierId, req.Name, req.IdCardNum, imageUrl, callAt, status, uniqueId, responseJSON, req.TraceId}
	case tps.AuthTypeFaceMatch:
		sql = fmt.Sprintf("INSERT INTO %s(tenant_id,supplier_id,image_url,call_at,status,score,unique_id,response_data,trace_id) VALUES(?,?,?,?,?,?,?,?,?)", tableName)
		args = []interface{}{req.TenantId, supplierId, imageUrl, callAt, status, score, uniqueId, responseJSON, req.TraceId}
	}

	if err := s.db.Exec(sql, args...).Error; err != nil {
		log.Printf("[AuthService] traceId=%s, save call record error: %v, table=%s", req.TraceId, err, tableName)
	}
}

func isPrepaid(pattern int) bool {
	return pattern == 1
}

// parseScoreToInt mirrors Java responseData.getIntValue("score") (truncates decimals).
func parseScoreToInt(v interface{}) int {
	switch vt := v.(type) {
	case int:
		return vt
	case int64:
		return int(vt)
	case float64:
		return int(vt)
	case string:
		if f, err := strconv.ParseFloat(vt, 64); err == nil {
			return int(f)
		}
		var i int
		fmt.Sscanf(vt, "%d", &i)
		return i
	default:
		return 0
	}
}

func extractAuthResponse(responseData map[string]interface{}) *tps.AuthResponse {
	if v, ok := responseData["authResponse"]; ok {
		if ar, ok := v.(*tps.AuthResponse); ok {
			return ar
		}
		if ar, ok := v.(tps.AuthResponse); ok {
			return &ar
		}
	}
	return nil
}

func logAuthCacheHit(req model.AuthRequest, authType int, success bool, note string) {
	log.Printf("[AuthService] traceId=%s cache hit tenantId=%s type=%d name=%s idCard=%s success=%v skipTPS=true%s",
		req.TraceId, req.TenantId, authType, req.Name, mask.MaskIDCard(req.IdCardNum), success, note)
}

func logAuthTPSResult(req model.AuthRequest, supplierId int64, authType int, success bool, responseData map[string]interface{}) {
	uniqueId := ""
	if v, ok := responseData["uniqueId"].(string); ok {
		uniqueId = v
	}
	log.Printf("[AuthService] traceId=%s tenantId=%s type=%d supplierId=%d name=%s idCard=%s tps success=%v uniqueId=%s",
		req.TraceId, req.TenantId, authType, supplierId, req.Name, mask.MaskIDCard(req.IdCardNum), success, uniqueId)
}

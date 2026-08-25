package tps

import (
	"encoding/json"
	"fmt"
)

// AuthType constants matching Java's AuthType.Algorithm enum
const (
	AuthTypeIdentityCard     = 1 // 二要素 (name + ID card)
	AuthTypeIdentityCardFace = 2 // 三要素 (name + ID card + face)
	AuthTypeFaceMatch        = 3 // 人脸比对
)

// TpsIdAuthEnum matches Java's TpsIdAuthEnum
const (
	ChannelChuangLan = 1
	ChannelBaiDu     = 2
	ChannelHuaWei    = 3
)

// AuthResponse matches Java's AuthResponse
type AuthResponse struct {
	IzSame bool   `json:"izSame"`
	Score  string `json:"score,omitempty"`
	Info   string `json:"info,omitempty"`
}

// MarshalJSON ensures Score defaults to "0" when empty, matching Java's default field initialization.
func (a AuthResponse) MarshalJSON() ([]byte, error) {
	type Alias AuthResponse
	aux := &struct {
		Score string `json:"score"`
		*Alias
	}{
		Alias: (*Alias)(&a),
	}
	if a.Score == "" {
		aux.Score = "0"
	} else {
		aux.Score = a.Score
	}
	return json.Marshal(aux)
}

// CommonTpsIdAuthParam matches Java's CommonTpsIdAuthParam
type CommonTpsIdAuthParam struct {
	TraceId   string
	AuthType  int
	Config    interface{} // JSON config from supplier
	Name      string
	IdCardNum string
	Image     string // base64 face image
}

// CommonTpsFaceMatchParam matches Java's CommonTpsFaceMatchParam
type CommonTpsFaceMatchParam struct {
	Config    interface{}
	Name      string
	IdCardNum string
	SourceImg string // base64
	TargetImg string // base64
	TraceId   string
}

// TpsIdAuth is the interface all TPS providers implement
type TpsIdAuth interface {
	Authenticate(param CommonTpsIdAuthParam, responseData map[string]interface{}) bool
	MatchFaceImage(param CommonTpsFaceMatchParam, responseData map[string]interface{}) bool
}

// GetTpsIdAuth returns the TPS provider by channel ID
func GetTpsIdAuth(channel int) TpsIdAuth {
	switch channel {
	case ChannelChuangLan:
		return &ChuangLanAuth{}
	case ChannelBaiDu:
		return &BaiDuAuth{}
	case ChannelHuaWei:
		return &HuaWeiAuth{}
	default:
		return nil
	}
}

// asString coerces an arbitrary JSON-decoded value into a string, matching
// fastjson's lenient getString() behavior (numbers/bools become their literal
// text, nil becomes "").
func asString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		// Render integral values without a trailing ".0".
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", t)
	}
}

// asObject returns v as a map if it is a JSON object, otherwise an empty map,
// matching Java's Optional.ofNullable(getJSONObject(...)).orElse(new JSONObject()).
func asObject(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

// parseConfig converts supplier config to a typed struct.
// Matches Java CommonTpsIdAuthParam.parseConfig handling String and JSONObject.
func parseConfig(config interface{}, target interface{}) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	switch c := config.(type) {
	case string:
		if c == "" {
			return fmt.Errorf("config is empty")
		}
		return json.Unmarshal([]byte(c), target)
	case []byte:
		if len(c) == 0 {
			return fmt.Errorf("config is empty")
		}
		return json.Unmarshal(c, target)
	case map[string]interface{}:
		data, err := json.Marshal(c)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, target)
	default:
		data, err := json.Marshal(config)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, target)
	}
}

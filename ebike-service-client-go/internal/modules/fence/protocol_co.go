package fence

import (
	"encoding/json"

	"ebike-service-client-go/internal/pkg/javacompat"
)

type configProtocolCO struct {
	Id         *javacompat.LongStr  `json:"id"`
	ServiceId  *javacompat.LongStr  `json:"serviceId"`
	Title      *string              `json:"title"`
	Type       *int                 `json:"type"`
	Content    *string              `json:"content"`
	UpdatedPin *string              `json:"updatedPin"`
	UpdatedAt  *javacompat.DateTime `json:"updatedAt"`
}

func convertConfigProtocolCO(raw json.RawMessage) json.RawMessage {
	if javacompat.IsNullJSON(raw) {
		return raw
	}
	var co configProtocolCO
	if json.Unmarshal(raw, &co) != nil {
		return raw
	}
	b, err := javacompat.MarshalJSONNoHTMLEscape(co)
	if err != nil {
		return raw
	}
	return b
}

package xiaoan

import (
	"ebike-device-openapi-go/internal/api/dto"
)

// Bin201ShuaKaMessage 刷卡中控上报消息体。
// 对应 Java feature/Bin201Decode-20240531 分支中的 Bin201ShuaKaMessage。
type Bin201ShuaKaMessage struct {
	Imei          *string `json:"imei,omitempty"`
	MsgType       string  `json:"msgType"`
	Cmd           int16   `json:"cmd"`
	BussinessType string  `json:"bussinessType"`

	// CardNo 卡号标识，报文固定 20 字节 US-ASCII，解码后按 Java trim() 去除首尾填充。
	CardNo string `json:"cardNo"`
	// AuthNo 身份证号，报文固定 18 字节 US-ASCII，解码后按 Java trim() 去除首尾填充。
	AuthNo string `json:"authNo"`
	// Timestamp 刷卡时间戳，uint32 大端无符号整数（Java 中为 long/readUnsignedInt）。
	Timestamp uint32 `json:"timestamp"`
}

// Bin201Decode 解析小安协议 cmd=201（刷卡中控）上行报文。
// 报文布局：cardNo(20B) + authNo(18B) + timestamp(4B)，共 42 字节。
type Bin201Decode struct{}

func (d *Bin201Decode) Decode(header *dto.MessageHeader, data *ByteBuf) (interface{}, error) {
	msg := &Bin201ShuaKaMessage{
		MsgType:       "data",
		Cmd:           201, // CmdConstant.CMD_SHUAKA
		BussinessType: "ebike",
	}

	// 与 Java Bin201Decode.doDecode 一致：readCharSequence 后调用 trim()。
	msg.CardNo = javaTrim(data.ReadCharSequence(20))
	msg.AuthNo = javaTrim(data.ReadCharSequence(18))
	msg.Timestamp = data.ReadUnsignedInt()

	return msg, nil
}

// javaTrim 复刻 Java String.trim() 语义：移除首尾所有码点 <= 0x20（U+0020）的字符，
// 包含空格、制表符、控制字符，以及定长字段常见的 \0 填充。
//
// 注意：不能用 strings.TrimSpace 替代——后者基于 unicode.IsSpace，不会移除 \0，
// 会导致 \0 填充的卡号/身份证号与 Java 结果不一致。
// 按字节比较是安全的：UTF-8 多字节字符的每个字节都 >= 0x80 > 0x20，不会被误删。
func javaTrim(s string) string {
	start := 0
	end := len(s)
	for start < end && s[start] <= ' ' {
		start++
	}
	for end > start && s[end-1] <= ' ' {
		end--
	}
	return s[start:end]
}

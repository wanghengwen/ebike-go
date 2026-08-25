package gateway

import (
	"strings"
)

func keepNBit(str string, bit int) string {
	if !strings.Contains(str, ".") {
		zeros := strings.Repeat("0", bit)
		return str + "." + zeros
	}
	arr := strings.Split(str, ".")
	if len(arr) < 2 || arr[1] == "" {
		zeros := strings.Repeat("0", bit)
		return arr[0] + "." + zeros
	}
	dec := arr[1]
	if bit > len(dec) {
		zeros := strings.Repeat("0", bit-len(dec))
		return arr[0] + "." + dec + zeros
	}

	lastDigitChar := dec[bit-1]
	lastDigit := int(lastDigitChar - '0')
	suffix := "0"
	if lastDigit > 4 {
		suffix = "5"
	}
	return arr[0] + "." + dec[0:bit-1] + suffix
}

func keep6Bit2Str(str string) string {
	if !strings.Contains(str, ",") {
		return str
	}
	arr := strings.Split(str, ",")
	return keepNBit(arr[0], 6) + "," + keepNBit(arr[1], 6)
}

package crypto

import (
	"bytes"
	"testing"
)

func TestBuildLuopingDeviceStr(t *testing.T) {
	got := BuildLuopingDeviceStr("CNT123456", "123456789012345")
	want := "CNT123456+luopingtech.com+123456789012345"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if len(got) != 41 {
		t.Fatalf("L=%d want 41", len(got))
	}
}

func TestDeriveLuopingKeyDocExamples(t *testing.T) {
	deviceStr := BuildLuopingDeviceStr("CNT123456", "123456789012345")
	key0, err := DeriveLuopingKey(deviceStr, 0)
	if err != nil {
		t.Fatal(err)
	}
	if string(key0) != "CNT123456+luopin" {
		t.Fatalf("keyIndex=0: %q", key0)
	}
	key10, err := DeriveLuopingKey(deviceStr, 10)
	if err != nil {
		t.Fatal(err)
	}
	// index 10 is 'l' of "luoping..." (doc example text miscounted the offset)
	if string(key10) != deviceStr[10:26] {
		t.Fatalf("keyIndex=10: %q want %q", key10, deviceStr[10:26])
	}
	// Wrap: L=41, index 30 → data[30:41]+data[0:5]
	key30, err := DeriveLuopingKey(deviceStr, 30)
	if err != nil {
		t.Fatal(err)
	}
	want30 := deviceStr[30:] + deviceStr[:5]
	if string(key30) != want30 {
		t.Fatalf("keyIndex=30: %q want %q", key30, want30)
	}
	if _, err := DeriveLuopingKey(deviceStr, 41); err == nil {
		t.Fatal("keyIndex >= L must be rejected")
	}
}

func TestDeriveLuopingKeyInvalidIndex(t *testing.T) {
	if _, err := DeriveLuopingKey("abc", 3); err == nil {
		t.Fatal("expected error for keyIndex >= L")
	}
}

func TestLuopingAESRoundTrip(t *testing.T) {
	deviceID := "2604000274"
	imei := "868480084776843"
	plain := []byte{0xAA, 0x55, 0x00, 0x01, 0x00, 0x05, 'h', 'e', 'l', 'l', 'o'}
	for _, idx := range []byte{0, 1, 10, 30} {
		enc, err := EncryptLuopingMQTTWithKeyIndex(plain, deviceID, imei, idx)
		if err != nil {
			t.Fatalf("encrypt idx=%d: %v", idx, err)
		}
		if enc[0] != idx {
			t.Fatalf("payload keyIndex=%d want %d", enc[0], idx)
		}
		dec, err := DecryptLuopingMQTT(enc, deviceID, imei)
		if err != nil {
			t.Fatalf("decrypt idx=%d: %v", idx, err)
		}
		if !bytes.Equal(dec, plain) {
			t.Fatalf("roundtrip mismatch idx=%d", idx)
		}
	}
}

func TestEncryptLuopingMQTTRandomKeyIndex(t *testing.T) {
	plain := []byte("aa55-frame")
	enc, err := EncryptLuopingMQTT(plain, "dev1", "imei1")
	if err != nil {
		t.Fatal(err)
	}
	dec, err := DecryptLuopingMQTT(enc, "dev1", "imei1")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dec, plain) {
		t.Fatalf("got %q want %q", dec, plain)
	}
}

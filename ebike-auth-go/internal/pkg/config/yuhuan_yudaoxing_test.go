package config

import "testing"

func TestApplyYuhuanYudaoxingCustomFromEnv(t *testing.T) {
	t.Run("YUHUAN_YUDAOXING_CUSTOM true", func(t *testing.T) {
		AppConfig.YuhuanYudaoxingCustom = false
		t.Setenv("YUHUAN_YUDAOXING_CUSTOM", "true")
		t.Setenv("AUTH_YUHUAN_YUDAOXING_CUSTOM", "false")
		applyYuhuanYudaoxingCustomFromEnv()
		if !AppConfig.YuhuanYudaoxingCustom {
			t.Fatal("expected YuhuanYudaoxingCustom true")
		}
	})

	t.Run("AUTH_YUHUAN_YUDAOXING_CUSTOM false", func(t *testing.T) {
		AppConfig.YuhuanYudaoxingCustom = true
		t.Setenv("YUHUAN_YUDAOXING_CUSTOM", "")
		t.Setenv("AUTH_YUHUAN_YUDAOXING_CUSTOM", "false")
		applyYuhuanYudaoxingCustomFromEnv()
		if AppConfig.YuhuanYudaoxingCustom {
			t.Fatal("expected YuhuanYudaoxingCustom false")
		}
	})
}

func TestParseBoolEnv(t *testing.T) {
	if got, ok := parseBoolEnv("yes"); !ok || !got {
		t.Fatalf("parseBoolEnv(yes) = %v, %v", got, ok)
	}
	if got, ok := parseBoolEnv("off"); !ok || got {
		t.Fatalf("parseBoolEnv(off) = %v, %v", got, ok)
	}
	if _, ok := parseBoolEnv("maybe"); ok {
		t.Fatal("expected invalid bool env")
	}
}

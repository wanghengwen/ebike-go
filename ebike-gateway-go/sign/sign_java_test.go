package sign

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

var (
	javaRefOnce sync.Once
	javaRefJar  string
	javaRefErr  error
)

func buildJavaRefJar() (string, error) {
	_, filename, _, _ := runtime.Caller(0)
	signDir := filepath.Dir(filename)
	projectRoot := filepath.Dir(signDir)
	jarPath := filepath.Join(projectRoot, "testdata", "signref", "target", "signref-1.0.0.jar")
	if _, err := os.Stat(jarPath); err != nil {
		cmd := exec.Command("mvn", "-q", "package", "-DskipTests")
		cmd.Dir = filepath.Join(projectRoot, "testdata", "signref")
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("build signref jar failed: %w\n%s", err, out)
		}
	}
	return jarPath, nil
}

func ensureJavaRefJar(t *testing.T) string {
	t.Helper()
	javaRefOnce.Do(func() {
		javaRefJar, javaRefErr = buildJavaRefJar()
	})
	if javaRefErr != nil {
		t.Fatalf("java reference jar unavailable: %v", javaRefErr)
	}
	return javaRefJar
}

func runJavaSignRef(t *testing.T, mode, signType, timestamp, secret, payload string) string {
	t.Helper()
	jar := ensureJavaRefJar(t)
	cmd := exec.Command("java", "-jar", jar, mode, signType, timestamp, secret, payload)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("java signref failed: %v\n%s", err, out)
	}
	return string(out)
}

func TestSignQueryMatchesJava(t *testing.T) {
	cases := []struct {
		name      string
		query     string
		timestamp string
		secret    string
	}{
		{name: "empty query", query: "", timestamp: "1710000000000", secret: "test-secret"},
		{name: "single param", query: "name=alice", timestamp: "1710000000000", secret: "test-secret"},
		{name: "sorted keys and duplicate values", query: "z=9&a=2&a=1&m=hello", timestamp: "1710000000123", secret: "s3cr3t-key"},
		{name: "special chars", query: "tenantId=1001&userName=test%40mail.com", timestamp: "1710000000456", secret: "gateway-secret"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			params, err := url.ParseQuery(tc.query)
			if err != nil {
				t.Fatalf("parse query: %v", err)
			}
			goSign := SignQuery(params, TimestampHeaderKey, tc.timestamp, tc.secret)
			javaSign := runJavaSignRef(t, "business", "get", tc.timestamp, tc.secret, tc.query)
			if goSign != javaSign {
				t.Fatalf("sign mismatch\ngo : %s\njava: %s\ndata: %q", goSign, javaSign, BuildQuerySignData(params, TimestampHeaderKey, tc.timestamp, tc.secret))
			}
		})
	}
}

func TestTrimJSONMatchesJava(t *testing.T) {
	cases := []struct {
		name string
		json string
	}{
		{name: "compact object", json: `{"b":2,"a":1}`},
		{name: "with null", json: `{"name":"bob","age":null,"active":true}`},
		{name: "nested", json: `{"user":{"id":1,"tags":["a","b"]},"count":0}`},
		{name: "pretty input", json: "{\n  \"x\" : 1,\n  \"y\" : null\n}"},
		{name: "base64 slash plus equals", json: `{"encryptedData":"abc/def+ghi==","iv":"xyz=="}`},
		{name: "version dots", json: `{"platform":"wechat","version":"1.0.9","userPin":"a_123"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			goTrim := TrimJSON(tc.json)
			javaTrim := runJavaSignRef(t, "business", "trim", "", "", tc.json)
			if goTrim != javaTrim {
				t.Fatalf("trim mismatch\ngo : %s\njava: %s", goTrim, javaTrim)
			}
		})
	}
}

func TestSignJSONMatchesJava(t *testing.T) {
	cases := []struct {
		name      string
		body      string
		timestamp string
		secret    string
	}{
		{name: "empty body", body: "", timestamp: "1710000000000", secret: "test-secret"},
		{name: "simple json", body: `{"orderId":"12345","amount":99.5}`, timestamp: "1710000000000", secret: "test-secret"},
		{name: "json with null", body: `{"name":"bob","age":null,"active":true}`, timestamp: "1710000000999", secret: "gateway-secret"},
		{name: "pretty json", body: "{\n  \"tenantId\" : \"1001\",\n  \"page\" : 1\n}", timestamp: "1710000000123", secret: "s3cr3t"},
		{
			name: "rent unFrozenOrder version field",
			body: `{"platform":"wechat","traceId":"178133727282095238428475","deviceId":"deviceId_178133513486411442429108","tenantId":"1003","userLat":0,"userLng":0,"userPin":"a_2a5611900380a75","version":"1.0.9"}`,
			timestamp: "1781337272820",
			secret:    "a3sw1aw4ij0zyueqd81l1zv4iryw4kcd8jy3uufs7csqo1trkr6zbuj6e32qx7mg",
		},
		{
			name: "rent getRideInfo float coords",
			body: `{"platform":"wechat","traceId":"17813372730845029568187","deviceId":"deviceId_178133513486411442429108","tenantId":"1003","userLat":33.76674072265625,"userLng":118.38504340277778,"userPin":"a_2a5611900380a75","version":"1.0.9"}`,
			timestamp: "1781337273084",
			secret:    "a3sw1aw4ij0zyueqd81l1zv4iryw4kcd8jy3uufs7csqo1trkr6zbuj6e32qx7mg",
		},
		{
			name: "oauth token base64 fields",
			body: `{"platform":"wechat","traceId":"178133729447761449943486","deviceId":"deviceId_178046433882594330682407","tenantId":"1007","grant_type":"wechat_miniapp","appId":"wxd456774aeef4b810","js_code":"0b1RygGa17UnTL02jyFa17NOtA0RygGE","encryptedData":"0iKnmr7voi/NrSyLWBxeBxY21NNq9td1UEYpbt4pbXE2kCkKIZvL/BBGr/kKt9P098MUyXiHaocUGjG4VdSepJ4BHWV+wHWbquJzc2T2sbctnmKeuiqULCWU8V1Sb6bi2TJPJ4zzpG4vi89x7mpICvb8fiEhf/lVj0OKCjYmz5U/K1A0Z+xqiSCA/OQtBj4GXYblJY9mEWQc+THG9rN8NQ==","iv":"h4cquDIWLH6NMWS1r6pK3Q=="}`,
			timestamp: "1781337294477",
			secret:    "a3sw1aw4ij0zyueqd81l1zv4iryw4kcd8jy3uufs7csqo1trkr6zbuj6e32qx7mg",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			goSign := SignJSON(tc.body, TimestampHeaderKey, tc.timestamp, tc.secret)
			javaSign := runJavaSignRef(t, "business", "json", tc.timestamp, tc.secret, tc.body)
			if goSign != javaSign {
				t.Fatalf("sign mismatch\ngo : %s\njava: %s\ndata: %q", goSign, javaSign, BuildJSONSignData(tc.body, TimestampHeaderKey, tc.timestamp, tc.secret))
			}
		})
	}
}

func TestSignFormMatchesJava(t *testing.T) {
	cases := []struct {
		name      string
		mode      string
		parseMode FormParseMode
		body      string
		timestamp string
		secret    string
	}{
		{name: "business empty value", mode: "business", parseMode: FormParseBusiness, body: "name=bob&flag&city=sh", timestamp: "1710000000000", secret: "test-secret"},
		{name: "business sorted", mode: "business", parseMode: FormParseBusiness, body: "z=9&a=2&a=1", timestamp: "1710000000123", secret: "s3cr3t"},
		{name: "client skip invalid pair", mode: "client", parseMode: FormParseClient, body: "name=bob&flag&city=sh", timestamp: "1710000000000", secret: "test-secret"},
		{name: "client valid pairs", mode: "client", parseMode: FormParseClient, body: "z=9&a=2&a=1", timestamp: "1710000000123", secret: "s3cr3t"},
		{name: "business value contains equals", mode: "business", parseMode: FormParseBusiness, body: "token=abc=def", timestamp: "1710000000999", secret: "s3cr3t"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			goSign := SignForm(tc.body, tc.parseMode, TimestampHeaderKey, tc.timestamp, tc.secret)
			javaSign := runJavaSignRef(t, tc.mode, "form", tc.timestamp, tc.secret, tc.body)
			if goSign != javaSign {
				t.Fatalf("sign mismatch\ngo : %s\njava: %s\ndata: %q", goSign, javaSign, BuildFormSignData(tc.body, tc.parseMode, TimestampHeaderKey, tc.timestamp, tc.secret))
			}
		})
	}
}

package middleware

import "testing"

func TestAccessLogRequestBody(t *testing.T) {
	cases := []struct {
		name        string
		contentType string
		body        []byte
		want        string
	}{
		{name: "empty", contentType: "application/json", body: nil, want: ""},
		{name: "json", contentType: "application/json", body: []byte(`{"a":1}`), want: `{"a":1}`},
		{name: "multipart", contentType: "multipart/form-data; boundary=abc", body: []byte("binary\xff\xd8\xff"), want: "[binary body omitted, 9 bytes]"},
		{name: "octet stream", contentType: "application/octet-stream", body: []byte{0x00, 0x01}, want: "[binary body omitted, 2 bytes]"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := accessLogRequestBody(tc.contentType, tc.body)
			if got != tc.want {
				t.Fatalf("accessLogRequestBody() = %q, want %q", got, tc.want)
			}
		})
	}
}

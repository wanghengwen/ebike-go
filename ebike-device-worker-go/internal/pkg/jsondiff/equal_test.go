package jsondiff

import "testing"

func TestEqual(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{`{"code":"0","data":{"id":123}}`, `{"code":"0","data":{"id":"123"}}`, true},
		{`{"count":10}`, `{"count":"10"}`, true},
		{`{"asc":true}`, `{"asc":false}`, false},
		{`{"x":1.5}`, `{"x":"1.5"}`, true},
		{`{"x":1.5}`, `{"x":"2"}`, false},
		{`[1,2,3]`, `[1,"2",3]`, true},
		{`{"lat":31.000001,"lng":121.000001}`, `{"lat":31.000002,"lng":121.000002}`, true},
		{`{"lat":31.0,"lng":121.0}`, `{"lat":34.0,"lng":121.0}`, true},
		{`{"lat":31.0,"lng":121.0}`, `{"lat":35.0,"lng":121.0}`, false},
		{`{"a":1,"b":null}`, `{"a":1}`, true},
		{`{"a":1}`, `{"a":1,"b":null}`, true},
		{
			`{"success":true,"data":[{"lat":22.97,"lng":115.35,"speed":0.0,"course":0.0,"timestamp":1781587449,"totalMiles":0}]}`,
			`{"success":true,"data":[{"lng":115.35,"lat":22.97,"timestamp":1781587449,"speed":0,"course":0,"totalMiles":0}]}`,
			true,
		},
		{
			`{"success":true,"data":{"1":[{"lat":22.97,"lng":115.35,"speed":0.0,"course":0.0,"timestamp":1,"totalMiles":null}]}}`,
			`{"success":true,"data":{"1":[{"lng":115.35,"lat":22.97,"timestamp":1,"speed":0,"course":0}]}}`,
			true,
		},
	}
	for _, tc := range cases {
		if got := Equal([]byte(tc.a), []byte(tc.b)); got != tc.want {
			t.Errorf("Equal(%s, %s) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

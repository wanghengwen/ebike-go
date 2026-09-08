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
	}
	for _, tc := range cases {
		if got := Equal([]byte(tc.a), []byte(tc.b)); got != tc.want {
			t.Errorf("Equal(%s, %s) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

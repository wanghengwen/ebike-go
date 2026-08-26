// Command logextract turns a production ebike-device-paas log (captured by the
// CommandContextAspect) into Go test fixtures.
//
// Each "log aspect" entry carries: url, req (valid JSON), rep (success/code/msg +
// a Java Lombok toString `data`), clazz, cusTime, errorMsg. This tool extracts
// every entry, converts the Lombok `data` into JSON, and writes one case file per
// request under an output dir grouped by endpoint.
//
// Usage:
//
//	go run ./tools/logextract -in <prod.log> -out testdata
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	ansiRe    = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	tsRe      = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)
	urlRe     = regexp.MustCompile(`url='([^']*)'`)
	clazzRe   = regexp.MustCompile(`clazz='([^']*)'`)
	cusTimeRe = regexp.MustCompile(`cusTime=(\d+)`)
	errMsgRe  = regexp.MustCompile(`errorMsg='(.*)'\}\s*$`)
	successRe = regexp.MustCompile(`"success":(true|false)`)
	codeRe    = regexp.MustCompile(`"code":"?([^",}]*)"?`)
	msgRe     = regexp.MustCompile(`"msg":"([^"]*)"`)
)

// Case is one extracted request/response fixture.
type Case struct {
	URL      string          `json:"url"`
	Clazz    string          `json:"clazz,omitempty"`
	CusTime  int             `json:"cusTime,omitempty"`
	Req      json.RawMessage `json:"req,omitempty"`
	Rep      json.RawMessage `json:"rep,omitempty"`
	ErrorMsg string          `json:"errorMsg,omitempty"`
}

func main() {
	in := flag.String("in", "", "input production log file")
	out := flag.String("out", "testdata", "output directory for fixtures")
	maxPer := flag.Int("max", 30, "max cases per endpoint")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "usage: logextract -in <log> -out <dir>")
		os.Exit(2)
	}

	data, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", *in, err)
		os.Exit(1)
	}
	clean := ansiRe.ReplaceAllString(string(data), "")
	records := splitRecords(clean)

	byEndpoint := map[string][]Case{}
	total, parsed := 0, 0
	for _, rec := range records {
		if !strings.Contains(rec, "log aspect: {") {
			continue
		}
		total++
		c, ok := parseRecord(rec)
		if !ok {
			continue
		}
		parsed++
		key := sanitize(c.URL)
		byEndpoint[key] = append(byEndpoint[key], c)
	}

	endpoints := make([]string, 0, len(byEndpoint))
	for k := range byEndpoint {
		endpoints = append(endpoints, k)
	}
	sort.Strings(endpoints)

	written := 0
	for _, ep := range endpoints {
		cases := byEndpoint[ep]
		if len(cases) > *maxPer {
			cases = cases[:*maxPer]
		}
		dir := filepath.Join(*out, ep)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", dir, err)
			continue
		}
		for i, c := range cases {
			b, _ := json.MarshalIndent(c, "", "  ")
			fp := filepath.Join(dir, fmt.Sprintf("%03d.json", i))
			if err := os.WriteFile(fp, b, 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "write %s: %v\n", fp, err)
				continue
			}
			written++
		}
		fmt.Printf("%-45s %d cases\n", ep, len(cases))
	}
	fmt.Printf("\naspect entries: %d, parsed: %d, files written: %d, endpoints: %d\n",
		total, parsed, written, len(endpoints))
}

// splitRecords joins multi-line (pretty-printed) records, starting a new record
// at each timestamp-prefixed line.
func splitRecords(s string) []string {
	lines := strings.Split(s, "\n")
	var records []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			records = append(records, cur.String())
			cur.Reset()
		}
	}
	for _, ln := range lines {
		if tsRe.MatchString(ln) {
			flush()
		}
		cur.WriteString(ln)
		cur.WriteString("\n")
	}
	flush()
	return records
}

func parseRecord(rec string) (Case, bool) {
	var c Case
	m := urlRe.FindStringSubmatch(rec)
	if m == nil {
		return c, false
	}
	c.URL = m[1]
	if m := clazzRe.FindStringSubmatch(rec); m != nil {
		c.Clazz = m[1]
	}
	if m := cusTimeRe.FindStringSubmatch(rec); m != nil {
		c.CusTime, _ = strconv.Atoi(m[1])
	}
	if m := errMsgRe.FindStringSubmatch(rec); m != nil && m[1] != "null" {
		c.ErrorMsg = m[1]
	}

	// req = balanced {...} after "req="
	if reqStr, ok := balancedAfter(rec, "req="); ok {
		if compact := compactJSON(reqStr); compact != nil {
			c.Req = compact
		}
	}

	// rep = "null" or balanced {...} after "rep="
	if idx := strings.Index(rec, "rep="); idx >= 0 {
		after := strings.TrimSpace(rec[idx+len("rep="):])
		if strings.HasPrefix(after, "null") {
			c.Rep = json.RawMessage("null")
		} else if repStr, ok := balancedAfter(rec, "rep="); ok {
			if repJSON := convertRep(repStr); repJSON != nil {
				c.Rep = repJSON
			}
		}
	}
	return c, true
}

// balancedAfter returns the balanced {...} block immediately following marker.
func balancedAfter(s, marker string) (string, bool) {
	idx := strings.Index(s, marker)
	if idx < 0 {
		return "", false
	}
	rest := s[idx+len(marker):]
	start := strings.IndexByte(rest, '{')
	if start < 0 {
		return "", false
	}
	depth := 0
	for i := start; i < len(rest); i++ {
		switch rest[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return rest[start : i+1], true
			}
		}
	}
	return "", false
}

// compactJSON validates and compacts a JSON string; returns nil on failure.
func compactJSON(s string) json.RawMessage {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

// convertRep turns a rep block {"success":..,"code":..,"msg":..,"data":<lombok>}
// into proper JSON, converting the Lombok toString `data` into a JSON value.
func convertRep(rep string) json.RawMessage {
	out := map[string]interface{}{}
	if m := successRe.FindStringSubmatch(rep); m != nil {
		out["success"] = m[1] == "true"
	}
	if m := codeRe.FindStringSubmatch(rep); m != nil {
		out["code"] = m[1]
	}
	if m := msgRe.FindStringSubmatch(rep); m != nil {
		out["msg"] = m[1]
	}
	if idx := strings.Index(rep, `"data":`); idx >= 0 {
		dataStr := strings.TrimSpace(rep[idx+len(`"data":`):])
		dataStr = strings.TrimSuffix(strings.TrimSpace(dataStr), "}")
		dataStr = strings.TrimSpace(dataStr)
		if dataStr != "" && dataStr != "null" {
			p := &lombok{s: dataStr}
			if v, ok := p.parseValue(); ok {
				out["data"] = v
			} else {
				out["data"] = dataStr // fallback: keep raw
			}
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil
	}
	return b
}

// lombok parses Java Lombok-style toString output:
//
//	Name(a=1, b=null, c=[], d=[1,2], e=Nested(x=1))
type lombok struct {
	s string
	i int
}

func (p *lombok) parseValue() (interface{}, bool) {
	p.skipSpace()
	if p.i >= len(p.s) {
		return nil, false
	}
	switch p.s[p.i] {
	case '[':
		return p.parseList()
	}
	// Object: IDENT '(' ... ')'
	if name, ok := p.peekIdentParen(); ok {
		_ = name
		return p.parseObject()
	}
	return p.parseScalar()
}

func (p *lombok) parseObject() (interface{}, bool) {
	// consume identifier
	for p.i < len(p.s) && isIdentChar(p.s[p.i]) {
		p.i++
	}
	if p.i >= len(p.s) || p.s[p.i] != '(' {
		return nil, false
	}
	p.i++ // '('
	obj := map[string]interface{}{}
	for {
		p.skipSpace()
		if p.i < len(p.s) && p.s[p.i] == ')' {
			p.i++
			break
		}
		// key
		ks := p.i
		for p.i < len(p.s) && p.s[p.i] != '=' && p.s[p.i] != ')' {
			p.i++
		}
		if p.i >= len(p.s) || p.s[p.i] != '=' {
			return obj, true
		}
		key := strings.TrimSpace(p.s[ks:p.i])
		p.i++ // '='
		v, ok := p.parseValue()
		if !ok {
			return obj, true
		}
		obj[key] = v
		p.skipSpace()
		if p.i < len(p.s) && p.s[p.i] == ',' {
			p.i++ // ','
			continue
		}
		if p.i < len(p.s) && p.s[p.i] == ')' {
			p.i++
			break
		}
	}
	return obj, true
}

func (p *lombok) parseList() (interface{}, bool) {
	p.i++ // '['
	list := []interface{}{}
	for {
		p.skipSpace()
		if p.i < len(p.s) && p.s[p.i] == ']' {
			p.i++
			break
		}
		v, ok := p.parseValue()
		if !ok {
			break
		}
		list = append(list, v)
		p.skipSpace()
		if p.i < len(p.s) && p.s[p.i] == ',' {
			p.i++
			continue
		}
		if p.i < len(p.s) && p.s[p.i] == ']' {
			p.i++
			break
		}
	}
	return list, true
}

// parseScalar reads until a top-level delimiter: ',' ')' ']'.
func (p *lombok) parseScalar() (interface{}, bool) {
	start := p.i
	for p.i < len(p.s) {
		ch := p.s[p.i]
		if ch == ',' || ch == ')' || ch == ']' {
			break
		}
		p.i++
	}
	raw := strings.TrimSpace(p.s[start:p.i])
	if raw == "null" {
		return nil, true
	}
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return n, true
	}
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		return f, true
	}
	if raw == "true" || raw == "false" {
		return raw == "true", true
	}
	return raw, true
}

// peekIdentParen reports whether the next token is IDENT followed by '('.
func (p *lombok) peekIdentParen() (string, bool) {
	j := p.i
	for j < len(p.s) && isIdentChar(p.s[j]) {
		j++
	}
	if j > p.i && j < len(p.s) && p.s[j] == '(' {
		return p.s[p.i:j], true
	}
	return "", false
}

func (p *lombok) skipSpace() {
	for p.i < len(p.s) && (p.s[p.i] == ' ' || p.s[p.i] == '\n' || p.s[p.i] == '\t' || p.s[p.i] == '\r') {
		p.i++
	}
}

func isIdentChar(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || b == '_' || b == '$' || b == '.'
}

func sanitize(url string) string {
	s := strings.TrimPrefix(url, "/")
	s = strings.ReplaceAll(s, "/", "_")
	if s == "" {
		s = "root"
	}
	return s
}

package javalog

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Entry is one CommandContextAspect log-aspect record from Java ebike-fence logs.
type Entry struct {
	URL     string          `json:"url"`
	Request json.RawMessage `json:"request,omitempty"`
	Reply   string          `json:"reply"`
	Source  string          `json:"source"`
}

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func extractBalanced(s string, start int, open, close byte) (string, int, bool) {
	if start >= len(s) || s[start] != open {
		return "", start, false
	}
	depth := 0
	inString := false
	escape := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inString {
			if escape {
				escape = false
				continue
			}
			if c == '\\' {
				escape = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		if c == '"' {
			inString = true
			continue
		}
		if c == open {
			depth++
		} else if c == close {
			depth--
			if depth == 0 {
				return s[start : i+1], i + 1, true
			}
		}
	}
	return "", start, false
}

// ParseFile extracts log-aspect entries from a Java ebike-fence log file.
func ParseFile(path string) ([]Entry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := stripANSI(string(raw))
	const marker = "log aspect: {url='"
	var out []Entry
	for {
		idx := strings.Index(text, marker)
		if idx < 0 {
			break
		}
		text = text[idx+len(marker):]
		urlEnd := strings.Index(text, "'")
		if urlEnd < 0 {
			break
		}
		url := text[:urlEnd]
		text = text[urlEnd+1:]

		reqIdx := strings.Index(text, "req=")
		if reqIdx < 0 {
			continue
		}
		text = text[reqIdx+4:]
		var reqJSON string
		if strings.HasPrefix(text, ", rep=") {
			reqJSON = ""
		} else if strings.HasPrefix(text, "{") {
			block, next, ok := extractBalanced(text, 0, '{', '}')
			if !ok {
				continue
			}
			reqJSON = block
			text = text[next:]
		} else {
			continue
		}

		repIdx := strings.Index(text, "rep=")
		if repIdx < 0 {
			continue
		}
		text = text[repIdx+4:]
		if !strings.HasPrefix(text, "{") {
			continue
		}
		repBlock, _, ok := extractBalanced(text, 0, '{', '}')
		if !ok {
			continue
		}

		var reqRaw json.RawMessage
		if reqJSON != "" && json.Valid([]byte(reqJSON)) {
			reqRaw = json.RawMessage(reqJSON)
		}
		out = append(out, Entry{
			URL:     url,
			Request: reqRaw,
			Reply:   repBlock,
		})
	}
	for i := range out {
		out[i].Source = path
		if i := strings.LastIndexAny(path, `/\`); i >= 0 {
			out[i].Source = path[i+1:]
		}
	}
	return out, nil
}

// ReplyEnvelope parses the outer JSON fields success/code/msg from a Java log rep string.
// Java logs embed data as toString(), e.g. ComputeDistanceCO(distance=1618.0, izCloseLine=false).
type ReplyEnvelope struct {
	Success bool            `json:"success"`
	Code    json.RawMessage `json:"code"`
	Msg     string          `json:"msg"`
}

func ParseReplyEnvelope(rep string) (ReplyEnvelope, error) {
	// Normalize Java data=Foo(...) to data=null for envelope parsing only.
	normalized := javaDataToNull(rep)
	var env ReplyEnvelope
	err := json.Unmarshal([]byte(normalized), &env)
	return env, err
}

var javaDataRe = regexp.MustCompile(`"data":\s*[A-Za-z0-9_]+\([^)]*\)`)

func javaDataToNull(rep string) string {
	return javaDataRe.ReplaceAllString(rep, `"data":null`)
}

func urlPathSegments(url string) []string {
	path := strings.ToLower(strings.SplitN(url, "?", 2)[0])
	return strings.Split(strings.Trim(path, "/"), "/")
}

func urlHasSegment(url, segment string) bool {
	segment = strings.ToLower(segment)
	for _, seg := range urlPathSegments(url) {
		if seg == segment {
			return true
		}
	}
	return false
}

// IsReadOnlyURL classifies endpoints that do not mutate persistent state in normal success paths.
func IsReadOnlyURL(url string) bool {
	lower := strings.ToLower(url)
	if urlHasSegment(url, "returncar") || urlHasSegment(url, "ridingcar") ||
		urlHasSegment(url, "bindparking") || urlHasSegment(url, "unbindparking") ||
		urlHasSegment(url, "addexposure") {
		return false
	}
	mutating := []string{
		"/create", "/update", "/delete", "/add", "/edit", "/del",
		"/bind", "/unbind", "/lock", "/unlock", "/enable", "/disable",
		"/copy", "/deal", "/save", "/ins", "/upd", "/recharge", "/sort",
		"/batch", "/setdefault", "/onoff", "/deregister",
		"/sethelmet", "/setpoint", "/settbeacon", "/setdirection",
		"/setkickstand", "/setcamera", "/setlocation", "/remove",
		"/addclick", "/sitesapplication", "/dealsiteapplication",
	}
	for _, k := range mutating {
		if strings.Contains(lower, k) {
			return false
		}
	}
	readHints := []string{
		"/get", "/query", "/list", "/page", "/compute", "/select", "/filter",
		"/check", "/detail", "/bytype", "/default", "/app", "/izcan",
		"/accunlock", "/helmetstate", "/getfencerelation", "/fencerelation",
	}
	for _, h := range readHints {
		if strings.Contains(lower, h) {
			return true
		}
	}
	return strings.Contains(lower, "checkpart")
}

// IsBareResponseURL reports API paths that return the payload directly without Result envelope.
// Java FenceTagApi.getAll returns List<FenceTagCO> (raw JSON array).
func IsBareResponseURL(url string) bool {
	switch url {
	case "/fence/tags/getAll":
		return true
	default:
		return false
	}
}

// BareResponseURLs lists paths checked by readonly API scripts for bare JSON array bodies.
func BareResponseURLs() []string {
	return []string{"/fence/tags/getAll"}
}

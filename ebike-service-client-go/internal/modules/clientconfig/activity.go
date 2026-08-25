package clientconfig

import "strings"

// javaSplit mirrors Java String.split(","): trailing empty strings removed;
// zero-length input yields [""].
func javaSplit(s string) []string {
	if !strings.Contains(s, ",") {
		return []string{s}
	}
	parts := strings.Split(s, ",")
	n := len(parts)
	for n > 0 && parts[n-1] == "" {
		n--
	}
	return parts[:n]
}

func containsAny(a, b []string) bool {
	set := make(map[string]struct{}, len(a))
	for _, s := range a {
		set[s] = struct{}{}
	}
	for _, s := range b {
		if _, ok := set[s]; ok {
			return true
		}
	}
	return false
}

func isShowActivity(byRegister, byTags bool, visible *int, isNewUser bool, userTags, activeTags string) bool {
	registerShow := false
	tagsShow := false
	if byRegister {
		registerShow = visible != nil && (*visible == 0 || (*visible == 1 && isNewUser))
	}
	if byTags {
		tagsShow = containsAny(javaSplit(activeTags), javaSplit(userTags))
	}
	return registerShow || tagsShow
}

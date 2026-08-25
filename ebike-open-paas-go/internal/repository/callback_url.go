package repository

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrInvalidCallbackURL marks a rejection caused by the URL itself, so the
// handler can answer with a parameter error instead of a server-side one.
var ErrInvalidCallbackURL = errors.New("invalid callback url")

// ValidateCallbackURL accepts only HTTPS callback endpoints.
//
// Scheme is the only restriction: customers may point at private IPs, cluster
// DNS names or URLs that embed credentials. HTTPS still guarantees the delivery
// body is not readable on the wire.
func ValidateCallbackURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("%w: url is empty", ErrInvalidCallbackURL)
	}
	if len(raw) > 1024 {
		return fmt.Errorf("%w: url is too long", ErrInvalidCallbackURL)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: not a valid URL", ErrInvalidCallbackURL)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("%w: scheme must be https", ErrInvalidCallbackURL)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("%w: url has no host", ErrInvalidCallbackURL)
	}
	return nil
}

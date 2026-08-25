package proxy

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"ebike-gateway-go/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type PredicateArgs struct {
	Pattern string `yaml:"pattern"`
}

type Predicate struct {
	Name string        `yaml:"name"`
	Args PredicateArgs `yaml:"args"`
}

type FilterArgs struct {
	Enabled    bool   `yaml:"enabled"`
	Secret     string `yaml:"secret"`
	IgnoreUrls string `yaml:"ignoreUrls"`
	Timeout    int    `yaml:"timeout"`
}

type Filter struct {
	Name string     `yaml:"name"`
	Args FilterArgs `yaml:"args"`
}

type SpringGatewayRoute struct {
	Id         string      `yaml:"id"`
	Uri        string      `yaml:"uri"`
	Predicates []Predicate `yaml:"predicates"`
	Filters    []Filter    `yaml:"filters"`
}

type Route struct {
	Id          string
	Uri         string
	PathPattern string
	PathRegexp  *regexp.Regexp
	TargetURL   *url.URL

	// Route specific configs
	SecurityEnabled bool
	SecuritySecret  string
	SecurityIgnores []string

	SignEnabled bool
	SignSecret  string
	SignTimeout int
	SignIgnores []string
}

func ParseRoutes(yamlContent string) ([]Route, error) {
	var springRoutes []SpringGatewayRoute
	if err := yaml.Unmarshal([]byte(yamlContent), &springRoutes); err != nil {
		return nil, fmt.Errorf("failed to parse routes yaml: %w", err)
	}

	var routes []Route
	for _, sr := range springRoutes {
		r := Route{
			Id:  sr.Id,
			Uri: sr.Uri,
		}

		// Parse Predicates
		for _, p := range sr.Predicates {
			if p.Name == "Path" {
				r.PathPattern = p.Args.Pattern
				if r.PathPattern != "" {
					pStr := strings.ReplaceAll(r.PathPattern, "**", "___DOUBLE_STAR___")
					pStr = strings.ReplaceAll(pStr, "*", "[^/]*")
					pStr = strings.ReplaceAll(pStr, "___DOUBLE_STAR___", ".*")
					rx, err := regexp.Compile("^" + pStr + "$")
					if err == nil {
						r.PathRegexp = rx
					} else {
						logger.Log.Error("Failed to compile route regex", zap.String("pattern", r.PathPattern), zap.Error(err))
					}
				}
			}
		}

		// Preparse Target URL
		targetUri := r.Uri
		if strings.HasPrefix(targetUri, "lb://") {
			serviceName := strings.TrimPrefix(targetUri, "lb://")
			targetUri = "http://" + serviceName
		}
		parsedURL, err := url.Parse(targetUri)
		if err == nil {
			r.TargetURL = parsedURL
		} else {
			logger.Log.Error("Failed to parse route target URL", zap.String("uri", targetUri), zap.Error(err))
		}

		// Parse Filters
		for _, f := range sr.Filters {
			if f.Name == "Security" {
				r.SecurityEnabled = f.Args.Enabled
				r.SecuritySecret = f.Args.Secret
				if f.Args.IgnoreUrls != "" {
					r.SecurityIgnores = strings.Split(f.Args.IgnoreUrls, ":")
				}
			} else if f.Name == "Sign" {
				r.SignEnabled = f.Args.Enabled
				r.SignSecret = f.Args.Secret
				r.SignTimeout = f.Args.Timeout
				if f.Args.IgnoreUrls != "" {
					r.SignIgnores = strings.Split(f.Args.IgnoreUrls, ":")
				}
			}
		}

		routes = append(routes, r)
	}

	var routeLog []string
	for _, r := range routes {
		routeLog = append(routeLog, fmt.Sprintf("%s(%s -> %s)", r.Id, r.PathPattern, r.Uri))
	}

	logger.Log.Info("Parsed dynamic routes successfully",
		zap.Int("count", len(routes)),
		zap.Strings("routes", routeLog),
	)
	return routes, nil
}

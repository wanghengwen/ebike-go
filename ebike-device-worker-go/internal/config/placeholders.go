package config

import "strings"

func isUnresolvedPlaceholder(value string) bool {
	return strings.Contains(value, "${")
}

func applicationName(c *Config) string {
	if c == nil {
		return "ebike-device-worker"
	}
	name := strings.TrimSpace(c.Spring.Application.Name)
	if name == "" || isUnresolvedPlaceholder(name) {
		return "ebike-device-worker"
	}
	return name
}

func resolveAppName(fastidAppName, springAppName string) string {
	appName := strings.TrimSpace(springAppName)
	if appName == "" || isUnresolvedPlaceholder(appName) {
		appName = "ebike-device-worker"
	}
	if fastidAppName == "" || isUnresolvedPlaceholder(fastidAppName) {
		return appName
	}
	resolved := strings.ReplaceAll(fastidAppName, "${spring.application.name}", appName)
	if isUnresolvedPlaceholder(resolved) {
		return appName
	}
	return resolved
}

func resolvePlaceholders(c *Config) {
	if c == nil {
		return
	}
	c.Spring.Application.Name = applicationName(c)
	c.Spring.Xyy.Fastid.AppName = resolveAppName(c.Spring.Xyy.Fastid.AppName, c.Spring.Application.Name)
}

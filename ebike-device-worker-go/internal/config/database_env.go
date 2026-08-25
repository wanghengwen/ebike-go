package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

// applyDatabaseEnvOverrides applies database settings after file/Nacos merge.
//
//   - DATABASE_DSN: replace entire connection string
//   - DATABASE_NAME: keep host/user/password/query from Nacos, only change database name
func applyDatabaseEnvOverrides(c *Config) {
	if c == nil {
		return
	}
	if v := strings.TrimSpace(os.Getenv("DATABASE_DSN")); v != "" {
		c.Database.Dsn = v
		log.Printf("[config] database dsn set from DATABASE_DSN")
		return
	}
	name := strings.TrimSpace(os.Getenv("DATABASE_NAME"))
	if name == "" {
		return
	}
	if c.Database.Dsn == "" {
		log.Printf("[WARN] DATABASE_NAME=%s ignored: database dsn is empty (load Nacos first?)", name)
		return
	}
	newDSN, err := replacePostgresDatabaseName(c.Database.Dsn, name)
	if err != nil {
		log.Printf("[WARN] DATABASE_NAME override failed: %v", err)
		return
	}
	c.Database.Dsn = newDSN
	log.Printf("[config] database name overridden to %s", strings.TrimPrefix(name, "/"))
}

func replacePostgresDatabaseName(dsn, dbName string) (string, error) {
	dbName = strings.TrimSpace(dbName)
	if dbName == "" {
		return "", fmt.Errorf("database name is empty")
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("parse dsn: %w", err)
	}
	parsed.Path = "/" + strings.TrimPrefix(dbName, "/")
	return parsed.String(), nil
}

// ReplacePostgresDatabaseNameForTest exposes DSN db-name rewrite for tests.
func ReplacePostgresDatabaseNameForTest(dsn, dbName string) (string, error) {
	return replacePostgresDatabaseName(dsn, dbName)
}

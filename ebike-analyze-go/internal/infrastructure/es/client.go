package es

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"go.uber.org/zap"
)

var Client *elasticsearch.Client

// Config holds ES connection configuration.
type Config struct {
	Hostname           string
	Username           string
	Password           string
	ConnectTimeout     int // milliseconds, default 5000
	SocketTimeout      int // milliseconds, default 20000
	TrackTotalHitsUpTo int // default 65535
}

// Init initializes the global ES client.
// Returns the config with defaults applied so callers can write them back to
// GlobalConfig (TrackTotalHitsUpTo etc. are consumed from GlobalConfig at query time).
func Init(cfg Config) (Config, error) {
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = 5000
	}
	if cfg.SocketTimeout == 0 {
		cfg.SocketTimeout = 20000
	}
	if cfg.TrackTotalHitsUpTo == 0 {
		cfg.TrackTotalHitsUpTo = 65535
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout: time.Duration(cfg.ConnectTimeout) * time.Millisecond,
		}).DialContext,
		ResponseHeaderTimeout: time.Duration(cfg.SocketTimeout) * time.Millisecond,
		MaxIdleConnsPerHost:   10,
	}

	esCfg := elasticsearch.Config{
		Addresses: []string{fmt.Sprintf("http://%s:9200", cfg.Hostname)},
		Transport: transport,
	}
	if cfg.Username != "" && cfg.Password != "" {
		esCfg.Username = cfg.Username
		esCfg.Password = cfg.Password
	}

	var err error
	Client, err = elasticsearch.NewClient(esCfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to create ES client: %w", err)
	}

	// Ping to verify connection
	res, err := Client.Info()
	if err != nil {
		zap.L().Warn("ES info request failed", zap.Error(err))
	} else {
		res.Body.Close()
		zap.L().Info("ES client initialized", zap.String("host", cfg.Hostname))
	}

	return cfg, nil
}

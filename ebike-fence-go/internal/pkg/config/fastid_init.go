package config

import (
	"log"

	"ebike-fence-go/internal/pkg/fastid"
)

// InitFastID configures the process-wide FastId generator (Java FenceKeyGenerator).
func InitFastID() {
	if !FastIDEnabled() {
		log.Println("[fastid] disabled; idgen uses local fallback")
		return
	}
	fc := ResolvedFastID()
	if fc.Secret == "" {
		log.Fatalf("fastid secret is required (Nacos fast_id.yaml or SPRING_XYY_FASTID_SECRET)")
	}
	if err := fastid.Init(fc); err != nil {
		log.Fatalf("fastid init failed: %v", err)
	}
	log.Printf("[fastid] initialized app=%s url=%s namespace=%s group=%s", fc.AppName, fc.URL, fc.Namespace, fc.GroupID)
}

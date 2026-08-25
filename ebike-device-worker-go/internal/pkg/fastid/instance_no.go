package fastid

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func resolveInstanceNo(cfg Config) (int64, error) {
	machineUUID := cfg.MachineUUID
	if machineUUID == "" {
		machineUUID = os.Getenv("fastid.machine.uuid")
	}
	if machineUUID == "" {
		ip := localIP()
		if ip == "" {
			ip = "127.0.0.1"
		}
		machineUUID = fmt.Sprintf("%s:%d", ip, cfg.Port)
	}

	var instanceNo *int64
	for attempt := 0; attempt <= 5; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Second)
			log.Printf("[fastid] retry %d fetching machine id", attempt)
		}
		no, err := fetchMachineID(cfg, machineUUID)
		if err != nil {
			log.Printf("[fastid] fetch machine id failed: %v", err)
			continue
		}
		instanceNo = &no
		break
	}

	dir := cfg.InstanceNoLocalDirectory
	if instanceNo != nil {
		if err := writeInstanceNo(dir, cfg.AppName, *instanceNo); err != nil {
			log.Printf("[fastid] write local instance no failed: %v", err)
		} else {
			log.Printf("[fastid] machine id=%d saved locally app=%s", *instanceNo, cfg.AppName)
		}
		return *instanceNo, nil
	}

	readNo, err := readInstanceNo(dir, cfg.AppName)
	if err != nil || readNo == nil {
		return 0, fmt.Errorf("fastid machine id unavailable and no local cache for app %s", cfg.AppName)
	}
	log.Printf("[fastid] using cached machine id=%d app=%s", *readNo, cfg.AppName)
	return *readNo, nil
}

func writeInstanceNo(dir, appName string, instanceNo int64) error {
	appDir := filepath.Join(dir, appName)
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(appDir, "id"), []byte(strconv.FormatInt(instanceNo, 10)), 0o644)
}

func readInstanceNo(dir, appName string) (*int64, error) {
	data, err := os.ReadFile(filepath.Join(dir, appName, "id"))
	if err != nil {
		return nil, err
	}
	text := string(data)
	if text == "" {
		return nil, fmt.Errorf("empty instance no file")
	}
	v, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

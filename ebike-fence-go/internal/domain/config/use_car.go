package config

// HelmetConfig mirrors Java HelmetConfig embedded in use-car config JSON.
type HelmetConfig struct {
	IzHelmetUnlock             *bool `json:"izHelmetUnlock"`
	IzHelmetRemovalDetection   *bool `json:"izHelmetRemovalDetection"`
	IzHelmetWearDetection      *bool `json:"izHelmetWearDetection"`
	IzRidingHelmetWear         *bool `json:"izRidingHelmetWear"`
	IzTempParkingReturnHelmet  *bool `json:"izTempParkingReturnHelmet"`
}

func (h *HelmetConfig) GetIzTempParkingReturnHelmet() bool {
	if h == nil || h.IzTempParkingReturnHelmet == nil {
		return false
	}
	return *h.IzTempParkingReturnHelmet
}

// ConfigUseCarCO mirrors Java ConfigUseCarCO (riding-car beacon gate).
type ConfigUseCarCO struct {
	ServiceId    *int64        `json:"serviceId"`
	IzBeacon     *bool         `json:"izBeacon"`
	NearLine     *int          `json:"nearLine"`
	HelmetConfig *HelmetConfig `json:"helmetConfig,omitempty"`
}

func (c *ConfigUseCarCO) GetIzBeacon() bool {
	if c == nil || c.IzBeacon == nil {
		return false
	}
	return *c.IzBeacon
}

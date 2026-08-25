package config

// ConfigBaseItemCO mirrors Java ConfigBaseItemCO (minimal fields for read paths).
type ConfigBaseItemCO struct {
	Id                   *int64 `json:"id,omitempty"`
	ServiceId            *int64 `json:"serviceId,omitempty"`
	SwapBatteryThreshold *int   `json:"swapBatteryThreshold,omitempty"`
	IzAutoSwapBattery    *bool  `json:"izAutoSwapBattery,omitempty"`
	IzHelmetAbnormalRiding *bool `json:"izHelmetAbnormalRiding,omitempty"`
}

func (c *ConfigBaseItemCO) GetIzHelmetAbnormalRiding() bool {
	if c == nil || c.IzHelmetAbnormalRiding == nil {
		return false
	}
	return *c.IzHelmetAbnormalRiding
}

package config

// ConfigBackcarCO represents the backend car configuration equivalent to ConfigBackcarCO in Java.
// All boolean flags are stored as pointers so we can correctly provide Java's default behaviors if null.
type ConfigBackcarCO struct {
	Id                         *int64   `json:"id"`
	AllowOutofService          *bool    `json:"allowOutofService"`
	AllowInNostop              *bool    `json:"allowInNostop"`
	AllowOutofParking          *bool    `json:"allowOutofParking"`
	AllowInBanRiding           *bool    `json:"allowInBanRiding"`
	DispatchCost               *int     `json:"dispatchCost"`
	PenaltyInNostop            *int     `json:"penaltyInNostop"`
	PenaltyOutofService        *int     `json:"penaltyOutofService"`
	PenaltyInBanRiding         *int     `json:"penaltyInBanRiding"`
	IzCivilizationRemind       *bool    `json:"izCivilizationRemind"`
	CoefficientOfDifficult     *float64 `json:"coefficientOfDifficult"`
	BufferDistance             *float64 `json:"bufferDistance"`
	ParkingBackcar             *int     `json:"parkingBackcar"`
	NoParkingBackcar           *int     `json:"noParkingBackcar"`
	BanRidingBackcar           *int     `json:"banRidingBackcar"`
	IzHelmetLock               *bool    `json:"izHelmetLock"`
	IzHelmetPhoto              *bool    `json:"izHelmetPhoto"`
	IzBeacon                   *bool    `json:"izBeacon"`
	ServiceId                  *int64   `json:"serviceId"`
	IzHelmetRemovalDetection   *bool    `json:"izHelmetRemovalDetection"`
	IzHelmetWearDetection      *bool    `json:"izHelmetWearDetection"`
	IzHelmetAutoRepair         *bool    `json:"izHelmetAutoRepair"`
	PowerOff                   *int     `json:"powerOff"`
	HelmetPenalty              *int     `json:"helmetPenalty"`
	IzHelmetReign              *bool    `json:"izHelmetReign"`
	IzCameraBackcar            *bool    `json:"izCameraBackcar"`
	IzCameraDirectionalBackcar *bool    `json:"izCameraDirectionalBackcar"`
	IzCameraPointBackcar       *bool    `json:"izCameraPointBackcar"`
	CameraImpunityCount        *int     `json:"cameraImpunityCount"`
	IzFullPileNoStop           *int     `json:"izFullPileNoStop"`
	IzUseOtherParking          *bool    `json:"izUseOtherParking"`
	BeaconEffect               *int     `json:"beaconEffect"`
	RfidEffect                 *int     `json:"rfidEffect"`
	DirectionEffect            *int     `json:"directionEffect"`
	FootSupportEffect          *int     `json:"footSupportEffect"`
	CameraEffect               *int     `json:"cameraEffect"`
	CameraAngle                *int     `json:"cameraAngle"`
	IzStandardReturn           *int     `json:"izStandardReturn"`
}

// Below are the safe getters replicating Java's null-safety fallbacks.

func (c *ConfigBackcarCO) GetIzHelmetReign() bool {
	if c.IzHelmetReign == nil {
		return true // default in Java
	}
	return *c.IzHelmetReign
}

func (c *ConfigBackcarCO) GetIzCameraBackcar() bool {
	if c.IzCameraBackcar == nil {
		return false
	}
	return *c.IzCameraBackcar
}

func (c *ConfigBackcarCO) GetIzUseOtherParking() bool {
	if c.IzUseOtherParking == nil {
		return false
	}
	return *c.IzUseOtherParking
}

func (c *ConfigBackcarCO) GetIzCameraDirectionalBackcar() bool {
	if c.IzCameraDirectionalBackcar == nil {
		return false
	}
	return *c.IzCameraDirectionalBackcar
}

func (c *ConfigBackcarCO) GetIzCameraPointBackcar() bool {
	if c.IzCameraPointBackcar == nil {
		return false
	}
	return *c.IzCameraPointBackcar
}

func (c *ConfigBackcarCO) GetCameraImpunityCount() int {
	if c.CameraImpunityCount == nil {
		return 4
	}
	return *c.CameraImpunityCount
}

func (c *ConfigBackcarCO) GetIzFullPileNoStop() int {
	if c.IzFullPileNoStop == nil {
		return 0
	}
	return *c.IzFullPileNoStop
}

func (c *ConfigBackcarCO) GetIzHelmetAutoRepair() bool {
	if c.IzHelmetAutoRepair == nil {
		return false
	}
	return *c.IzHelmetAutoRepair
}

func (c *ConfigBackcarCO) GetBufferDistance() float64 {
	if c.BufferDistance == nil {
		return 10.0 // Default from ConfigBackcarServiceImpl.java line 32
	}
	return *c.BufferDistance
}

func (c *ConfigBackcarCO) GetIzHelmetWearDetection() bool {
	if c == nil || c.IzHelmetWearDetection == nil {
		return false
	}
	return *c.IzHelmetWearDetection
}

func (c *ConfigBackcarCO) GetIzHelmetRemovalDetection() bool {
	if c == nil || c.IzHelmetRemovalDetection == nil {
		return false
	}
	return *c.IzHelmetRemovalDetection
}

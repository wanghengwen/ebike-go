package dto

import (
	"encoding/json"

	"ebike-fence-go/internal/domain/geo"
)

// FenceCO mirrors Java FenceCO base fields.
type FenceCO struct {
	Id                int64   `json:"id"`
	Type              int     `json:"type"`
	Name              string  `json:"name"`
	ShapeType         string  `json:"shapeType"`
	CenterLat         float64 `json:"centerLat"`
	CenterLng         float64 `json:"centerLng"`
	PointList         string  `json:"pointList"`
	Pics              string  `json:"pics"`
	TenantId          string  `json:"tenantId"`
	CreatedPin        string  `json:"createdPin"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedPin        string  `json:"updatedPin"`
	UpdatedAt         string  `json:"updatedAt"`
	OpeningHoursBegin string  `json:"openingHoursBegin"`
	OpeningHoursEnd   string  `json:"openingHoursEnd"`
}

type ServiceAreaCO struct {
	FenceCO
	DataVersion int64 `json:"dataVersion"`
}

type ServiceAreaIDCO struct {
	Id       int64  `json:"id"`
	TenantId string `json:"tenantId"`
}

type ServiceAreaCmd struct {
	Command
	Id        *int64  `json:"id,omitempty"`
	Name      string  `json:"name"`
	ShapeType string  `json:"shapeType"`
	CenterLat float64 `json:"centerLat"`
	CenterLng float64 `json:"centerLng"`
	PointList string  `json:"pointList"`
	Pics      string  `json:"pics,omitempty"`
}

type ComputeDistanceCmd struct {
	Command
	Imei      string       `json:"imei"`
	Point     *LocationCmd `json:"point"`
	PointJSON string       `json:"-"`
}

// UnmarshalJSON accepts Java-style point string or object {"lng","lat"}.
func (c *ComputeDistanceCmd) UnmarshalJSON(data []byte) error {
	type alias ComputeDistanceCmd
	var raw struct {
		alias
		Point json.RawMessage `json:"point"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*c = ComputeDistanceCmd(raw.alias)
	if len(raw.Point) == 0 {
		return nil
	}
	loc, err := geo.ParsePointJSON(raw.Point)
	if err != nil {
		c.PointJSON = string(raw.Point)
		return nil
	}
	c.Point = &LocationCmd{Lng: loc.Lng, Lat: loc.Lat}
	return nil
}

type ComputeDistanceCO struct {
	Distance   float64 `json:"distance"`
	IzCloseLine bool   `json:"izCloseLine"`
}

type TenantServiceCmd struct {
	Command
	RoleIds     []int64 `json:"roleIds"`
	SubTenantId string  `json:"subTenantId"`
}

type ServiceAreaLocationCmd struct {
	Command
	Lat *float64 `json:"lat"`
	Lng *float64 `json:"lng"`
}

type ParkingPartCO struct {
	RfidNum      int `json:"rfidNum"`
	DirectionNum int `json:"directionNum"`
	CameraNum    int `json:"cameraNum"`
	KickstandNum int `json:"kickstandNum"`
	TbeaconNum   int `json:"tbeaconNum"`
}

type ParkingCO struct {
	FenceCO
	MaxParkingNumber       int      `json:"maxParkingNumber"`
	ServiceId              int64    `json:"serviceId"`
	Tbeacon                *bool    `json:"tbeacon"`
	Directional            *bool    `json:"directional"`
	Direction              *float64 `json:"direction"`
	FormulateDirection     *float64 `json:"formulateDirection"`
	Rfid                   *bool    `json:"rfid"`
	IzEnable               bool     `json:"izEnable"`
	CoefficientOfDifficult *float64 `json:"coefficientOfDifficult"`
	BufferDistance         *float64 `json:"bufferDistance"`
	Camera                 *bool    `json:"camera"`
	IzCameraDirectionalBackcar *bool `json:"izCameraDirectionalBackcar"`
	IzCameraPointBackcar       *bool `json:"izCameraPointBackcar"`
	Kickstand              *bool    `json:"kickstand"`
	IzFullPileNoStop       *int     `json:"izFullPileNoStop"`
	AreaSize               *float64 `json:"areaSize"`
	UseCount               int      `json:"useCount"`
	OfflineCount           int      `json:"offlineCount"`
	RepairCount            int      `json:"repairCount"`
	OrderCount             int      `json:"orderCount"`
	OrderCost              int      `json:"orderCost"`
	FullCar                *bool    `json:"fullCar"`
	RefTags                []int64  `json:"refTags"`
	CarCount               *int64   `json:"carCount"`
	Distance               float64  `json:"distance"`
	OpeningHoursBegin      *string  `json:"openingHoursBegin"`
	OpeningHoursEnd        *string  `json:"openingHoursEnd"`
	IzOpenAllDay           *bool    `json:"izOpenAllDay"`
	ActivityId             *int64   `json:"activityId"`
	MinAmount              *int     `json:"minAmount"`
	MaxAmount              *int     `json:"maxAmount"`
	ExpirationTime         *string  `json:"expirationTime"`
	Rfids                  *string  `json:"rfids"`
}

type ParkingCmd struct {
	Command
	ParkingCO
}

type IdPageCmd struct {
	Command
	Id        *int64    `json:"id"`
	Name      string    `json:"name,omitempty"`
	AreaSize  *float64  `json:"areaSize,omitempty"`
	FieldList []string  `json:"fieldList,omitempty"`
	PageNum   int       `json:"pageNum,omitempty"`
	PageSize  int       `json:"pageSize,omitempty"`
}

type CarIdCmd struct {
	Command
	Id string `json:"id"`
}

type NearLocationCmd struct {
	Command
	ServiceId int64   `json:"serviceId"`
	Location  struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"locationCmd"`
	IzClient *bool    `json:"izClient,omitempty"`
	Radius   *float64 `json:"radius,omitempty"`
}

type LocationsCmd struct {
	Command
	ServiceId int64 `json:"serviceId"`
	Locations []struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"locations"`
}

type CreateParkingBatchCmd struct {
	Command
	ServiceId      int64        `json:"serviceId"`
	ParkingCmdList []ParkingCmd `json:"parkingCmdList"`
}

type UpdateParkingBatchCMD struct {
	Command
	Ids []int64 `json:"ids"`
	ParkingCmd
}

type ParkingBindCarCmd struct {
	Command
	Imei string   `json:"imei" binding:"required"`
	Lat  *float64 `json:"lat" binding:"required"`
	Lng  *float64 `json:"lng" binding:"required"`
}

type BindParkingCO struct {
	ParkingId      *int64 `json:"parkingId"`
	NoParkingId    *int64 `json:"noParkingId"`
	BanRidingId    *int64 `json:"banRidingId"`
	MaintainAreaId *int64 `json:"maintainAreaId"`
	NearParkingId  *int64 `json:"nearParkingId"`
	FenceCustomId  *int64 `json:"fenceCustomId"`
}

type ParkingUnBindCarCmd struct {
	Command
	Imei      string  `json:"imei"`
	ParkingId *int64  `json:"parkingId,omitempty"`
	CarId     *string `json:"carId"` // Java optional: if provided, skip imei→carId RPC
}

type ParkingPageQuery struct {
	Command
	ServiceId int64  `json:"serviceId"`
	Name      string `json:"name,omitempty"`
	Id        *int64 `json:"id,omitempty"`
	PageNum   int    `json:"pageNum,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
}

type NearParkingCO struct {
	IzOutService int     `json:"izOutService"`
	ParkingNum   int     `json:"parkingNum"`
	Distance     float64 `json:"distance"`
}

type ParkingCarCount struct {
	ParkingId        int64 `json:"parkingId"`
	CarCount         int64 `json:"carCount"`
	MaxParkingNumber int   `json:"maxParkingNumber"`
}

type CopyByServiceCmd struct {
	Command
	OriginalServiceId  int64   `json:"originalServiceId"`
	OriginalParkingIds []int64 `json:"originalParkingIds,omitempty"`
	TargetServiceId    int64   `json:"targetServiceId"`
}

type NoParkingCO struct {
	FenceCO
	ServiceId int64 `json:"serviceId"`
	Distance  float64 `json:"distance"`
}

type NoParkingCmd struct {
	Command
	NoParkingCO
}

type BanRidingCO struct {
	FenceCO
	ServiceId int64   `json:"serviceId"`
	Distance  float64 `json:"distance"`
}

type BanRidingCmd struct {
	Command
	BanRidingCO
}

type MaintainAreaCO struct {
	FenceCO
	ServiceId        int64   `json:"serviceId"`
	AreaSize         float64 `json:"areaSize"`
	MaxParkingNumber int     `json:"maxParkingNumber"`
	LeaderName       string  `json:"leaderName"`
	MaintainNum      int     `json:"maintainNum"`
}

type MaintainAreaCmd struct {
	Command
	MaintainAreaCO
}

type MaintainAreaPageQuery struct {
	Command
	ServiceId int64 `json:"serviceId"`
	PageNum   int   `json:"pageNum,omitempty"`
	PageSize  int   `json:"pageSize,omitempty"`
}

type AreaEmployeeVO struct {
	UserPin  string `json:"userPin"`
	Name     string `json:"name,omitempty"`
	Phone    string `json:"phone,omitempty"`
	RoleName string `json:"roleName,omitempty"`
	Type     int    `json:"type"`
}

type PersonnelDetailCO struct {
	LeaderList            []AreaEmployeeVO `json:"leaderList"`
	CheckerList           []AreaEmployeeVO `json:"checkerList"`
	MaintainerList        []AreaEmployeeVO `json:"maintainerList"`
	ChangePowerPersonList []AreaEmployeeVO `json:"changePowerPersonList"`
	MoveCarPersonList     []AreaEmployeeVO `json:"moveCarPersonList"`
}

type PersonnelEditCMD struct {
	Command
	AreaId int64            `json:"areaId"`
	List   []AreaEmployeeVO `json:"list"`
}

type AreaEmployeeCmd struct {
	Command
	UserPin       string `json:"userPin"`
	ServiceAreaId int64  `json:"serviceAreaId"`
	Type          *int   `json:"type,omitempty"`
}

type FenceCustomTypeCO struct {
	Id             int64  `json:"id"`
	Name           string `json:"name"`
	Color          string `json:"color"`
	Description    string `json:"description"`
	IzCreateCarTag bool   `json:"izCreateCarTag"`
	CarTagList     string `json:"carTagList"`
}

type FenceCustomTypeCmd struct {
	Command
	FenceCustomTypeCO
}

type FenceCustomCO struct {
	FenceCO
	ServiceId    int64  `json:"serviceId"`
	CustomTypeId int64  `json:"customTypeId"`
	Color        string `json:"color"`
	CarCount     int64  `json:"carCount"`
}

type FenceCustomCmd struct {
	Command
	FenceCustomCO
}

type CustomFenceListQry struct {
	Command
	ServiceId    int64 `json:"serviceId"`
	CustomTypeId int64 `json:"customTypeId,omitempty"`
}

type FenceRfidCO struct {
	Id       int64  `json:"id"`
	FenceId  int64  `json:"fenceId"`
	RfidCode string `json:"rfidCode"`
	IzDel    *int   `json:"izDel"`
}

type FenceRfidCmd struct {
	Command
	FenceId int64  `json:"fenceId"`
	IzDel   *int   `json:"izDel,omitempty"`
}

type RfidCmd struct {
	Command
	RfidCode string `json:"rfidCode"`
}

type FenceRfidSaveCmd struct {
	Command
	FenceId   int64    `json:"fenceId"`
	RfidCode  string   `json:"rfidCode,omitempty"`
	RfidCodes []string `json:"rfidCodes,omitempty"`
	IzDel     *int     `json:"izDel,omitempty"`
}

type FenceRfidChangeCmd struct {
	Command
	OldRfidCode string `json:"oldRfidCode"`
	NewRfidCode string `json:"newRfidCode"`
}

type FenceInfoByRfidsCmd struct {
	Command
	RfidCodes []string `json:"rfidCodes"`
}

type FenceInfoByRfidsCO struct {
	FenceId   int64                  `json:"fenceId"`
	FenceInfo map[string]interface{} `json:"fenceInfo"`
	RfidCode  string                 `json:"rfidCode"`
}

type FenceCmd struct {
	Command
	Id        int64  `json:"id"`
	ServiceId *int64 `json:"serviceId,omitempty"`
	IzEnable  *bool  `json:"izEnable,omitempty"`
}

type SiteApplicationCmd struct {
	Command
	Id              *int64  `json:"id,omitempty"`
	Phone           string  `json:"phone,omitempty"`
	ServiceId       int64   `json:"serviceId"`
	Location        string  `json:"location"`
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	PhotoUrl        string  `json:"photoUrl,omitempty"`
	ApplicantRemark string  `json:"applicantRemark"`
	State           *int    `json:"state,omitempty"`
	OpManRemark     string  `json:"opManRemark"`
	RemindWay       []int   `json:"remindWay,omitempty"`
}

type SiteApplicationCO struct {
	Id              int64  `json:"id"`
	ServiceId       int64  `json:"serviceId"`
	UserPin         string `json:"userPin"`
	Location        string `json:"location"`
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	ApplicantPhone  string `json:"applicantPhone"`
	PhotoUrl        string `json:"photoUrl"`
	ApplicantRemark string `json:"applicantRemark"`
	State           int    `json:"state"`
	OpManPin        string `json:"opManPin"`
	OpManPhone      string `json:"opManPhone"`
	OpManRemark     string `json:"opManRemark"`
	CreatedAt       string `json:"createdAt"`
}

type ApplicationQuery struct {
	Command
	ServiceId        int64  `json:"serviceId"`
	Phone            string `json:"phone,omitempty"`
	State            *int   `json:"state,omitempty"`
	OpManPhone       string `json:"opManPhone,omitempty"`
	CreatedTimeStart string `json:"createdTimeStart,omitempty"`
	CreatedTimeEnd   string `json:"createdTimeEnd,omitempty"`
	PageNum          int    `json:"pageNum,omitempty"`
	PageSize         int    `json:"pageSize,omitempty"`
	SearchCount      bool   `json:"searchCount,omitempty"`
}

type CreateRadPacketParkingCmd struct {
	Command
	List []struct {
		Id             int64  `json:"id"`
		ExpirationTime string `json:"expirationTime"`
	} `json:"list"`
	MinAmount  int   `json:"minAmount"`
	MaxAmount  int   `json:"maxAmount"`
	ActivityId int64 `json:"activityId"`
}

type ParkingStationFilterCmd struct {
	Command
	Id               int64    `json:"id"`
	ParkingId        *int64   `json:"parkingId,omitempty"`
	MaintainAreaId   *int64   `json:"maintainAreaId,omitempty"`
	Name             string   `json:"name,omitempty"`
	IzEnable         *int     `json:"izEnable,omitempty"`
	FieldList        []string `json:"fieldList"`
	Tags             []int    `json:"tags"`
	AreaSize         *float64 `json:"areaSize,omitempty"`
	UpdatedStartTime *string  `json:"updatedStartTime,omitempty"`
	UpdatedEndTime   *string  `json:"updatedEndTime,omitempty"`
	UpdatedPin       string   `json:"updatedPin,omitempty"`
	PageNum          int      `json:"pageNum,omitempty"`
	PageSize         int      `json:"pageSize,omitempty"`
	Orders           any      `json:"orders"`
	SearchCount      bool     `json:"searchCount"`
	LastRecordId     *string  `json:"lastRecordId,omitempty"`
}

type ParkingPageOrderByQry struct {
	Command
	ServiceId      int64 `json:"serviceId"`
	MaintainAreaId *int64 `json:"maintainAreaId,omitempty"`
	PageNum        int   `json:"pageNum,omitempty"`
	PageSize       int   `json:"pageSize,omitempty"`
}

type ParkingMonitorPageQry struct {
	Command
	ServiceId   int64  `json:"serviceId"`
	Level       *int   `json:"level,omitempty"`
	FieldList   string `json:"fieldList,omitempty"`
	IzFullPile  *bool  `json:"izFullPile,omitempty"`
	MinCarCount *int   `json:"minCarCount,omitempty"`
	MaxCarCount *int   `json:"maxCarCount,omitempty"`
	PageNum     int    `json:"pageNum,omitempty"`
	PageSize    int    `json:"pageSize,omitempty"`
}

type ParkingMonitorCO struct {
	ParkingCO
	Level int `json:"level"`
}

type ParkingMonitorDetailCO struct {
	ParkingCO
	CarCount int64 `json:"carCount"`
}

type ParkStationCarsPageQuery struct {
	Command
	ServiceId int64  `json:"serviceId"`
	ParkingId int64  `json:"parkingId"`
	Name      string `json:"name,omitempty"`
	Id        *int64 `json:"id,omitempty"`
	PageNum   int    `json:"pageNum,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
}

type ParkingDetailCO struct {
	CarId     string `json:"carId"`
	ParkingId int64  `json:"parkingId"`
}
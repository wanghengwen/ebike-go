package device

import (
	"strings"

	"ebike-device-worker-go/internal/domain/constants"
	"ebike-device-worker-go/internal/domain/entity"
	"ebike-device-worker-go/internal/domain/message"
	"ebike-device-worker-go/internal/infrastructure/persistence/pg"
	"ebike-device-worker-go/internal/pkg/fastid"
	"ebike-device-worker-go/internal/pkg/utils"
)

type gpsService struct{ repo *pg.Repo }
type pingService struct{ repo *pg.Repo }
type bmsService struct{ repo *pg.Repo }
type alarmService struct{ repo *pg.Repo }
type faultService struct{ repo *pg.Repo }
type cmdService struct{ repo *pg.Repo }
type onlineService struct{ repo *pg.Repo }
type offlineService struct{ repo *pg.Repo }

func NewGPS(repo *pg.Repo) Service     { return &gpsService{repo: repo} }
func NewPing(repo *pg.Repo) Service    { return &pingService{repo: repo} }
func NewBMS(repo *pg.Repo) Service     { return &bmsService{repo: repo} }
func NewAlarm(repo *pg.Repo) Service   { return &alarmService{repo: repo} }
func NewFault(repo *pg.Repo) Service   { return &faultService{repo: repo} }
func NewCmd(repo *pg.Repo) Service     { return &cmdService{repo: repo} }
func NewOnline(repo *pg.Repo) Service  { return &onlineService{repo: repo} }
func NewOffline(repo *pg.Repo) Service { return &offlineService{repo: repo} }

func (s *gpsService) Key() string     { return "ebike_" + utils.Itoa(constants.CmdGPS1) }
func (s *pingService) Key() string    { return "ebike_" + utils.Itoa(constants.CmdPing) }
func (s *bmsService) Key() string     { return "ebike_" + utils.Itoa(constants.CmdBMSInfo) }
func (s *alarmService) Key() string   { return "ebike_" + utils.Itoa(constants.CmdAlarm) }
func (s *faultService) Key() string   { return "ebike_" + utils.Itoa(constants.CmdFault) }
func (s *cmdService) Key() string     { return "ebike_" + utils.Itoa(constants.CmdWild) }
func (s *onlineService) Key() string  { return "ebike_" + utils.Itoa(constants.CmdLogin) }
func (s *offlineService) Key() string { return "ebike_" + utils.Itoa(constants.CmdLogout) }

func (s *gpsService) SaveBatch(list []message.MessageInfoDTO, batchSize int) error {
	var rows []pg.GpsRow
	for _, msg := range list {
		imei, metric, ts, geom, createTime, ok := entity.StructEBikeGps(msg)
		if !ok || !entity.GPSValidTimestamp(ts) {
			continue
		}
		rows = append(rows, pg.GpsRow{ID: fastid.Next(), Imei: imei, Metric: metric, Timestamp: ts, Geometry: geom, CreateTime: createTime})
	}
	return s.repo.InsertGpsBatch(rows, batchSize)
}

func (s *pingService) SaveBatch(list []message.MessageInfoDTO, batchSize int) error {
	var rows []pg.PingRow
	for _, msg := range list {
		imei, metric, ts, ok := entity.StructEBikePing(msg)
		if !ok {
			continue
		}
		rows = append(rows, pg.PingRow{ID: fastid.Next(), Imei: imei, Metric: metric, Timestamp: ts})
	}
	return s.repo.InsertPingBatch(rows, batchSize)
}

func (s *bmsService) SaveBatch(list []message.MessageInfoDTO, batchSize int) error {
	var rows []pg.BmsRow
	for _, msg := range list {
		imei, metric, ts, ok := entity.StructEBikeBms(msg)
		if !ok {
			continue
		}
		rows = append(rows, pg.BmsRow{ID: fastid.Next(), Imei: imei, Metric: metric, Timestamp: ts})
	}
	return s.repo.InsertBmsBatch(rows, batchSize)
}

func (s *alarmService) SaveBatch(list []message.MessageInfoDTO, batchSize int) error {
	var rows []pg.AlarmRow
	for _, msg := range list {
		imei, alarmType, ts, ok := entity.StructEBikeAlarm(msg)
		if !ok {
			continue
		}
		rows = append(rows, pg.AlarmRow{ID: fastid.Next(), Imei: imei, Type: alarmType, Timestamp: ts})
	}
	return s.repo.InsertAlarmBatch(rows, batchSize)
}

func (s *faultService) SaveBatch(list []message.MessageInfoDTO, batchSize int) error {
	var rows []pg.FaultRow
	for _, msg := range list {
		imei, etc, bms, ecu, bit, ts, ok := entity.StructEBikeFault(msg)
		if !ok {
			continue
		}
		rows = append(rows, pg.FaultRow{ID: fastid.Next(), Imei: imei, EtcFault: etc, BmsFault: bms, EcuFault: ecu, BitDetail: bit, Timestamp: ts})
	}
	return s.repo.InsertFaultBatch(rows, batchSize)
}

func (s *cmdService) SaveBatch(list []message.MessageInfoDTO, batchSize int) error {
	var rows []pg.CmdRow
	for _, msg := range list {
		imei, msgID, dir, cmdJSON, ts, ok := entity.StructEBikeCmd(msg)
		if !ok || len(msgID) != 32 {
			continue
		}
		rows = append(rows, pg.CmdRow{ID: fastid.Next(), Imei: imei, MsgID: msgID, Dir: dir, Cmd: cmdJSON, Timestamp: ts})
	}
	return s.repo.InsertCmdBatch(rows, batchSize)
}

func (s *onlineService) SaveBatch(list []message.MessageInfoDTO, batchSize int) error {
	return saveOnOffLine(s.repo, list, batchSize)
}
func (s *offlineService) SaveBatch(list []message.MessageInfoDTO, batchSize int) error {
	return saveOnOffLine(s.repo, list, batchSize)
}

func saveOnOffLine(repo *pg.Repo, list []message.MessageInfoDTO, batchSize int) error {
	var rows []pg.OnOffLineRow
	for _, msg := range list {
		imei, lineState, ts, ok := entity.StructEBikeOnOffLine(msg)
		if !ok {
			continue
		}
		rows = append(rows, pg.OnOffLineRow{ID: fastid.Next(), Imei: imei, LineState: lineState, Timestamp: ts})
	}
	return repo.InsertOnOffLineBatch(rows, batchSize)
}

// SanitizeStringFields mirrors Java getSendData string trim for \u0080.
func SanitizeStringFields(m map[string]interface{}) {
	for k, v := range m {
		if s, ok := v.(string); ok {
			m[k] = strings.TrimSpace(strings.ReplaceAll(s, "\u0080", ""))
		}
	}
}

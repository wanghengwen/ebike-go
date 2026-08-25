package pg

import (
	"fmt"
	"strings"
	"time"

	"ebike-device-worker-go/internal/pkg/db"

	"gorm.io/gorm"
)

type Repo struct{}

func NewRepo() *Repo { return &Repo{} }

type GpsRow struct {
	ID         int64
	Imei       string
	Metric     string
	Geometry   string
	Timestamp  time.Time
	CreateTime time.Time
}
type PingRow struct {
	ID        int64
	Imei      string
	Metric    string
	Timestamp time.Time
}
type BmsRow struct {
	ID        int64
	Imei      string
	Metric    string
	Timestamp time.Time
}
type AlarmRow struct {
	ID        int64
	Imei      string
	Type      int
	Timestamp time.Time
}
type FaultRow struct {
	ID        int64
	Imei      string
	EtcFault  int
	BmsFault  int
	EcuFault  int
	BitDetail string
	Timestamp time.Time
}
type CmdRow struct {
	ID        int64
	Imei      string
	MsgID     string
	Dir       string
	Cmd       string
	Timestamp time.Time
}
type OnOffLineRow struct {
	ID        int64
	Imei      string
	LineState string
	Timestamp time.Time
}

func (r *Repo) InsertGpsBatch(rows []GpsRow, batchSize int) error {
	return execInChunks(rows, batchSize, insertGpsChunk)
}

func insertGpsChunk(tx *gorm.DB, rows []GpsRow) error {
	ph := make([]string, len(rows))
	args := make([]interface{}, 0, len(rows)*6)
	for i, row := range rows {
		ph[i] = "(?,?,?::jsonb,?,ST_GeomFromText(?,4326),?)"
		args = append(args, row.ID, row.Imei, row.Metric, row.Timestamp, row.Geometry, row.CreateTime)
	}
	sql := fmt.Sprintf(`INSERT INTO %s(id, imei, metric, timestamp, geometry, create_time) VALUES `, db.PgTable("ebike_gps")) + strings.Join(ph, ",")
	return tx.Exec(sql, args...).Error
}

func (r *Repo) InsertPingBatch(rows []PingRow, batchSize int) error {
	return execInChunks(rows, batchSize, insertPingChunk)
}

func insertPingChunk(tx *gorm.DB, rows []PingRow) error {
	ph := make([]string, len(rows))
	args := make([]interface{}, 0, len(rows)*4)
	for i, row := range rows {
		ph[i] = "(?,?,?::jsonb,?)"
		args = append(args, row.ID, row.Imei, row.Metric, row.Timestamp)
	}
	sql := fmt.Sprintf(`INSERT INTO %s(id, imei, metric, timestamp) VALUES `, db.PgTable("ebike_ping")) + strings.Join(ph, ",")
	return tx.Exec(sql, args...).Error
}

func (r *Repo) InsertBmsBatch(rows []BmsRow, batchSize int) error {
	return execInChunks(rows, batchSize, insertBmsChunk)
}

func insertBmsChunk(tx *gorm.DB, rows []BmsRow) error {
	ph := make([]string, len(rows))
	args := make([]interface{}, 0, len(rows)*4)
	for i, row := range rows {
		ph[i] = "(?,?,?::jsonb,?)"
		args = append(args, row.ID, row.Imei, row.Metric, row.Timestamp)
	}
	sql := fmt.Sprintf(`INSERT INTO %s(id, imei, metric, timestamp) VALUES `, db.PgTable("ebike_bms")) + strings.Join(ph, ",")
	return tx.Exec(sql, args...).Error
}

func (r *Repo) InsertAlarmBatch(rows []AlarmRow, batchSize int) error {
	return execInChunks(rows, batchSize, insertAlarmChunk)
}

func insertAlarmChunk(tx *gorm.DB, rows []AlarmRow) error {
	ph := make([]string, len(rows))
	args := make([]interface{}, 0, len(rows)*4)
	for i, row := range rows {
		ph[i] = "(?,?,?,?)"
		args = append(args, row.ID, row.Imei, row.Type, row.Timestamp)
	}
	sql := fmt.Sprintf(`INSERT INTO %s(id, imei, type, timestamp) VALUES `, db.PgTable("ebike_alarm")) + strings.Join(ph, ",")
	return tx.Exec(sql, args...).Error
}

func (r *Repo) InsertFaultBatch(rows []FaultRow, batchSize int) error {
	return execInChunks(rows, batchSize, insertFaultChunk)
}

func insertFaultChunk(tx *gorm.DB, rows []FaultRow) error {
	ph := make([]string, len(rows))
	args := make([]interface{}, 0, len(rows)*7)
	for i, row := range rows {
		ph[i] = "(?,?,?,?,?,?::jsonb,?)"
		args = append(args, row.ID, row.Imei, row.EtcFault, row.BmsFault, row.EcuFault, row.BitDetail, row.Timestamp)
	}
	sql := fmt.Sprintf(`INSERT INTO %s(id, imei, etc_fault, bms_fault, ecu_fault, bit_detail, timestamp) VALUES `, db.PgTable("ebike_fault")) + strings.Join(ph, ",")
	return tx.Exec(sql, args...).Error
}

func (r *Repo) InsertCmdBatch(rows []CmdRow, batchSize int) error {
	return execInChunks(rows, batchSize, insertCmdChunk)
}

func insertCmdChunk(tx *gorm.DB, rows []CmdRow) error {
	ph := make([]string, len(rows))
	args := make([]interface{}, 0, len(rows)*6)
	for i, row := range rows {
		ph[i] = "(?,?,?,?::jsonb,?,?)"
		args = append(args, row.ID, row.Imei, row.Timestamp, row.Cmd, row.Dir, row.MsgID)
	}
	sql := fmt.Sprintf(`INSERT INTO %s(id, imei, timestamp, cmd, dir, msg_id) VALUES `, db.PgTable("ebike_cmd")) + strings.Join(ph, ",")
	return tx.Exec(sql, args...).Error
}

func (r *Repo) InsertOnOffLineBatch(rows []OnOffLineRow, batchSize int) error {
	return execInChunks(rows, batchSize, insertOnOffLineChunk)
}

func insertOnOffLineChunk(tx *gorm.DB, rows []OnOffLineRow) error {
	ph := make([]string, len(rows))
	args := make([]interface{}, 0, len(rows)*4)
	for i, row := range rows {
		ph[i] = "(?,?,?,?)"
		args = append(args, row.ID, row.Imei, row.LineState, row.Timestamp)
	}
	sql := fmt.Sprintf(`INSERT INTO %s(id, imei, line_state, timestamp) VALUES `, db.PgTable("ebike_onoffline")) + strings.Join(ph, ",")
	return tx.Exec(sql, args...).Error
}

func execInChunks[T any](rows []T, batchSize int, fn func(*gorm.DB, []T) error) error {
	if db.DB == nil || len(rows) == 0 {
		return nil
	}
	if batchSize <= 0 {
		batchSize = 20
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(rows); start += batchSize {
			end := start + batchSize
			if end > len(rows) {
				end = len(rows)
			}
			if err := fn(tx, rows[start:end]); err != nil {
				return err
			}
		}
		return nil
	})
}

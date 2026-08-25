package service

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"identity-auth-go/internal/model"
	goredis "identity-auth-go/internal/pkg/redis"

	"gorm.io/gorm"
)

type QueryService struct {
	db *gorm.DB
}

func NewQueryService(db *gorm.DB) *QueryService {
	return &QueryService{db: db}
}

// AuthRecordRow is the curated call-record projection used by /queryAuthRecord
// and /exportAuthRecord. It mirrors the exact field set that Java's
// CallRecordServiceImpl.query() builds (tenantId, name, identityNo, callAt,
// status, responseData); other columns are intentionally omitted.
type AuthRecordRow struct {
	TenantId     string
	Name         string
	IdentityNo   string
	CallAt       time.Time
	Status       int
	ResponseData string // raw JSON string from response_data column
}

// getTableName returns the dynamic table name based on auth type and date.
// Java uses: t_two_call_record_{yyyy_MM} for type=1, t_three_call_record_{yyyy_MM} for type=2
func getTableName(authType int, date string) (string, error) {
	switch authType {
	case 1: // IdentityCard (二要素)
		return "t_two_call_record_" + date, nil
	case 2: // IdentityCardFace (三要素)
		return "t_three_call_record_" + date, nil
	default:
		return "", fmt.Errorf("unsupported auth type: %d", authType)
	}
}

// CountAuthTimes returns {"used": X, "unused": Y} matching Java's countCallRecord.
// "used" comes from the DB count, "unused" comes from Redis (charge balance).
func (s *QueryService) CountAuthTimes(req model.AuthRecordRequest) (map[string]int, error) {
	tableName, err := getTableName(req.Type, req.Date)
	if err != nil {
		return nil, err
	}

	// Build SQL matching Java: SELECT count(1) as times FROM {table} [WHERE tenant_id = ?]
	sqlStr := fmt.Sprintf("SELECT count(1) as times FROM %s", tableName)
	var args []interface{}
	if req.TenantId != "" {
		sqlStr += " WHERE tenant_id = ?"
		args = append(args, req.TenantId)
	}

	var used int
	row := s.db.Raw(sqlStr, args...).Row()
	if err := row.Scan(&used); err != nil {
		return nil, err
	}

	// Get unused count from Redis (matching Java's countCallRecord Mono.zip)
	unused := 0
	if req.TenantId != "" {
		redisUnused, err := goredis.GetCallTimes(req.TenantId, req.Type)
		if err != nil {
			log.Printf("[QueryService] redis error getting unused count: %v", err)
		} else {
			unused = int(redisUnused)
		}
	}
	log.Printf("[QueryService] countAuthTimes: table=%s, tenantId=%s, used=%d, unused=%d, traceId=%s",
		tableName, req.TenantId, used, unused, req.TraceId)

	return map[string]int{"used": used, "unused": unused}, nil
}

// QueryAuthRecord queries call records from the dynamically-named table.
// Matches Java's query(): SELECT only the fields it maps, filter by status < 3.
func (s *QueryService) QueryAuthRecord(req model.AuthRecordRequest) ([]AuthRecordRow, error) {
	tableName, err := getTableName(req.Type, req.Date)
	if err != nil {
		return nil, err
	}

	// Java SQL: select * from {table} where status < 3 [and tenant_id = '...'].
	// We select only the columns Java maps to keep the response shape identical.
	sqlStr := fmt.Sprintf("SELECT tenant_id, name, identity_no, call_at, status, response_data FROM %s WHERE status < 3", tableName)
	var args []interface{}
	if req.TenantId != "" {
		sqlStr += " AND tenant_id = ?"
		args = append(args, req.TenantId)
	}

	log.Printf("[QueryService] queryAuthRecord: sql=%s, traceId=%s", sqlStr, req.TraceId)

	rows, err := s.db.Raw(sqlStr, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AuthRecordRow
	for rows.Next() {
		var (
			tenantID   sql.NullString
			name       sql.NullString
			identityNo sql.NullString
			callAtVal  interface{}
			status     int
			respData   sql.NullString
		)
		if err := rows.Scan(&tenantID, &name, &identityNo, &callAtVal, &status, &respData); err != nil {
			return nil, err
		}
		results = append(results, AuthRecordRow{
			TenantId:     tenantID.String,
			Name:         name.String,
			IdentityNo:   identityNo.String,
			CallAt:       toTime(callAtVal),
			Status:       status,
			ResponseData: respData.String,
		})
	}

	return results, nil
}

// toTime converts a scanned call_at value (time.Time, []byte or string,
// depending on the driver/DSN) into a time.Time in the local zone, matching
// Java's use of the system default time zone.
func toTime(v interface{}) time.Time {
	switch t := v.(type) {
	case time.Time:
		return t.In(time.Local)
	case []byte:
		return parseDBTime(string(t))
	case string:
		return parseDBTime(t)
	default:
		return time.Time{}
	}
}

func parseDBTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	layouts := []string{
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t
		}
	}
	return time.Time{}
}

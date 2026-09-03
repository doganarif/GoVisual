package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/doganarif/govisual/v2/store"
	_ "github.com/lib/pq"
)

// cleanupEveryN runs the capacity-trim query once every N successful inserts,
// instead of on every Add. Trading a slight overshoot of the configured capacity
// for far less load on the database.
const cleanupEveryN = 32

// Store implements the Store interface with PostgreSQL as backend
type Store struct {
	db          *sql.DB
	tableName   string
	capacity    int
	insertCount atomic.Uint64
}

// NewStore creates a new PostgreSQL-backed store
func New(connStr, tableName string, capacity int) (*Store, error) {
	if capacity <= 0 {
		capacity = 100
	}

	if !store.IsValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name %q: must match [A-Za-z_][A-Za-z0-9_]*", tableName)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	s := &Store{
		db:        db,
		tableName: tableName,
		capacity:  capacity,
	}

	if err := s.createTable(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return s, nil
}

func (s *Store) createTable() error {
	// `extras` holds fields added in v2 (logs, panic stack, performance
	// metrics) as a single JSONB blob so future capture fields don't need
	// schema changes.
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			timestamp TIMESTAMP WITH TIME ZONE,
			method TEXT,
			host TEXT,
			path TEXT,
			raw_path TEXT,
			query TEXT,
			request_headers JSONB,
			response_headers JSONB,
			status_code INTEGER,
			duration BIGINT,
			request_body TEXT,
			response_body TEXT,
			error TEXT,
			middleware_trace JSONB,
			route_trace JSONB,
			extras JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`, s.tableName)

	if _, err := s.db.Exec(query); err != nil {
		return err
	}

	// Upgrade existing tables one column at a time. IF NOT EXISTS makes this
	// safe when multiple application instances initialize concurrently.
	migrations := []struct {
		name       string
		definition string
	}{
		{name: "extras", definition: "JSONB"},
		{name: "host", definition: "TEXT"},
		{name: "raw_path", definition: "TEXT"},
	}
	for _, migration := range migrations {
		alter := fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s", s.tableName, migration.name, migration.definition)
		if _, err := s.db.Exec(alter); err != nil {
			return fmt.Errorf("add %s column: %w", migration.name, err)
		}
	}

	indexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s_timestamp_idx ON %s (timestamp DESC)",
		s.tableName, s.tableName)
	_, err := s.db.Exec(indexQuery)
	return err
}

// extrasPayload holds capture fields added in v2 that share a single JSONB
// column. Serialized to `extras` on write and unpacked on read.
type extrasPayload struct {
	Logs               []store.LogEntry          `json:"logs,omitempty"`
	PanicStack         string                    `json:"panic_stack,omitempty"`
	PerformanceMetrics *store.PerformanceMetrics `json:"performance_metrics,omitempty"`
}

func encodeExtras(l *store.RequestLog) (string, error) {
	p := extrasPayload{Logs: l.Logs, PanicStack: l.PanicStack, PerformanceMetrics: l.PerformanceMetrics}
	if len(p.Logs) == 0 && p.PanicStack == "" && p.PerformanceMetrics == nil {
		return "{}", nil
	}
	data, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("marshal extras: %w", err)
	}
	return string(data), nil
}

func decodeExtras(s string, l *store.RequestLog) {
	if s == "" || s == "{}" {
		return
	}
	var p extrasPayload
	if err := json.Unmarshal([]byte(s), &p); err != nil {
		log.Printf("govisual: unmarshaling extras for %s: %v", l.ID, err)
		return
	}
	l.Logs = p.Logs
	l.PanicStack = p.PanicStack
	l.PerformanceMetrics = p.PerformanceMetrics
}

// Add adds a new request log to the store
func (s *Store) Add(reqLog *store.RequestLog) error {
	reqHeaders, err := prepareJSON(reqLog.RequestHeaders)
	if err != nil {
		return fmt.Errorf("marshal request headers: %w", err)
	}
	respHeaders, err := prepareJSON(reqLog.ResponseHeaders)
	if err != nil {
		return fmt.Errorf("marshal response headers: %w", err)
	}

	middlewareTrace := "[]"
	if len(reqLog.MiddlewareTrace) > 0 {
		data, err := json.Marshal(reqLog.MiddlewareTrace)
		if err != nil {
			return fmt.Errorf("marshal middleware trace: %w", err)
		}
		middlewareTrace = string(data)
	}

	routeTrace := "{}"
	if reqLog.RouteTrace != nil {
		data, err := json.Marshal(reqLog.RouteTrace)
		if err != nil {
			return fmt.Errorf("marshal route trace: %w", err)
		}
		routeTrace = string(data)
	}

	extras, err := encodeExtras(reqLog)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, timestamp, method, host, path, raw_path, query, request_headers, response_headers,
			status_code, duration, request_body, response_body, error,
			middleware_trace, route_trace, extras
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9::jsonb, $10, $11, $12, $13, $14, $15::jsonb, $16::jsonb, $17::jsonb)
	`, s.tableName)

	_, err = s.db.Exec(
		query,
		reqLog.ID,
		reqLog.Timestamp,
		reqLog.Method,
		reqLog.Host,
		reqLog.Path,
		reqLog.RawPath,
		reqLog.Query,
		reqHeaders,
		respHeaders,
		reqLog.StatusCode,
		reqLog.Duration,
		reqLog.RequestBody,
		reqLog.ResponseBody,
		reqLog.Error,
		middlewareTrace,
		routeTrace,
		extras,
	)
	if err != nil {
		return fmt.Errorf("postgres insert: %w", err)
	}

	if s.insertCount.Add(1)%cleanupEveryN == 0 {
		if err := s.cleanup(); err != nil {
			return err
		}
	}
	return nil
}

// prepareJSON ensures we have a valid JSON string
func prepareJSON(v interface{}) (string, error) {
	if v == nil {
		return "{}", nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// cleanup removes old logs to maintain the capacity limit
func (s *Store) cleanup() error {
	// One statement that keeps the newest rows; a separate COUNT would go
	// stale under concurrent inserts and leave the table above capacity.
	deleteQuery := fmt.Sprintf(`
		DELETE FROM %s
		WHERE id NOT IN (
			SELECT id FROM %s
			ORDER BY timestamp DESC
			LIMIT $1
		)
	`, s.tableName, s.tableName)

	if _, err := s.db.Exec(deleteQuery, s.capacity); err != nil {
		return fmt.Errorf("postgres cleanup: %w", err)
	}
	return nil
}

// Get retrieves a specific request log by its ID
func (s *Store) Get(id string) (*store.RequestLog, bool) {
	query := fmt.Sprintf(`
		SELECT
			id, timestamp, method, COALESCE(host, ''), path, COALESCE(raw_path, ''), query,
			COALESCE(request_headers::text, '{}'),
			COALESCE(response_headers::text, '{}'),
			status_code, duration, request_body, response_body, error,
			COALESCE(middleware_trace::text, '[]'),
			COALESCE(route_trace::text, '{}'),
			COALESCE(extras::text, '{}')
		FROM %s
		WHERE id = $1
	`, s.tableName)

	var (
		reqLog          store.RequestLog
		reqHeadersStr   string
		respHeadersStr  string
		middlewareTrace string
		routeTrace      string
		extras          string
	)

	err := s.db.QueryRow(query, id).Scan(
		&reqLog.ID,
		&reqLog.Timestamp,
		&reqLog.Method,
		&reqLog.Host,
		&reqLog.Path,
		&reqLog.RawPath,
		&reqLog.Query,
		&reqHeadersStr,
		&respHeadersStr,
		&reqLog.StatusCode,
		&reqLog.Duration,
		&reqLog.RequestBody,
		&reqLog.ResponseBody,
		&reqLog.Error,
		&middlewareTrace,
		&routeTrace,
		&extras,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		log.Printf("govisual: failed to get request log from PostgreSQL: %v", err)
		return nil, false
	}

	unmarshalLogJSON(reqHeadersStr, &reqLog.RequestHeaders, "request_headers", reqLog.ID)
	unmarshalLogJSON(respHeadersStr, &reqLog.ResponseHeaders, "response_headers", reqLog.ID)
	unmarshalLogJSON(middlewareTrace, &reqLog.MiddlewareTrace, "middleware_trace", reqLog.ID)
	unmarshalLogJSON(routeTrace, &reqLog.RouteTrace, "route_trace", reqLog.ID)
	decodeExtras(extras, &reqLog)

	return &reqLog, true
}

// GetAll returns all stored request logs
func (s *Store) GetAll() []*store.RequestLog {
	query := fmt.Sprintf(`
		SELECT
			id, timestamp, method, COALESCE(host, ''), path, COALESCE(raw_path, ''), query,
			COALESCE(request_headers::text, '{}'),
			COALESCE(response_headers::text, '{}'),
			status_code, duration, request_body, response_body, error,
			COALESCE(middleware_trace::text, '[]'),
			COALESCE(route_trace::text, '{}'),
			COALESCE(extras::text, '{}')
		FROM %s
		ORDER BY timestamp DESC
	`, s.tableName)

	return s.queryLogs(query)
}

// GetLatest returns the n most recent request logs
func (s *Store) GetLatest(n int) []*store.RequestLog {
	query := fmt.Sprintf(`
		SELECT
			id, timestamp, method, COALESCE(host, ''), path, COALESCE(raw_path, ''), query,
			COALESCE(request_headers::text, '{}'),
			COALESCE(response_headers::text, '{}'),
			status_code, duration, request_body, response_body, error,
			COALESCE(middleware_trace::text, '[]'),
			COALESCE(route_trace::text, '{}'),
			COALESCE(extras::text, '{}')
		FROM %s
		ORDER BY timestamp DESC
		LIMIT $1
	`, s.tableName)

	return s.queryLogs(query, n)
}

func (s *Store) queryLogs(query string, args ...interface{}) []*store.RequestLog {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("govisual: failed to query logs from PostgreSQL: %v", err)
		return nil
	}
	defer rows.Close()

	var logs []*store.RequestLog

	for rows.Next() {
		var (
			reqLog          store.RequestLog
			reqHeadersStr   string
			respHeadersStr  string
			middlewareTrace string
			routeTrace      string
			extras          string
		)

		if err := rows.Scan(
			&reqLog.ID,
			&reqLog.Timestamp,
			&reqLog.Method,
			&reqLog.Host,
			&reqLog.Path,
			&reqLog.RawPath,
			&reqLog.Query,
			&reqHeadersStr,
			&respHeadersStr,
			&reqLog.StatusCode,
			&reqLog.Duration,
			&reqLog.RequestBody,
			&reqLog.ResponseBody,
			&reqLog.Error,
			&middlewareTrace,
			&routeTrace,
			&extras,
		); err != nil {
			log.Printf("govisual: failed to scan row: %v", err)
			continue
		}

		unmarshalLogJSON(reqHeadersStr, &reqLog.RequestHeaders, "request_headers", reqLog.ID)
		unmarshalLogJSON(respHeadersStr, &reqLog.ResponseHeaders, "response_headers", reqLog.ID)
		unmarshalLogJSON(middlewareTrace, &reqLog.MiddlewareTrace, "middleware_trace", reqLog.ID)
		unmarshalLogJSON(routeTrace, &reqLog.RouteTrace, "route_trace", reqLog.ID)
		decodeExtras(extras, &reqLog)

		logs = append(logs, &reqLog)
	}

	if err := rows.Err(); err != nil {
		log.Printf("govisual: error iterating over rows: %v", err)
	}

	return logs
}

// Clear clears all stored request logs
func (s *Store) Clear() error {
	query := fmt.Sprintf("TRUNCATE TABLE %s", s.tableName)
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("failed to clear logs: %w", err)
	}
	return nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// unmarshalLogJSON is shared by all SQL stores so they all report unmarshal
// errors consistently instead of silently dropping fields.
func unmarshalLogJSON(s string, v interface{}, field, logID string) {
	if s == "" {
		return
	}
	if err := json.Unmarshal([]byte(s), v); err != nil {
		log.Printf("govisual: failed to unmarshal %s for log %s: %v", field, logID, err)
	}
}

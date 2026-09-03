package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync/atomic"

	"github.com/doganarif/govisual/v2/store"
)

// cleanupEveryN runs the capacity trim once every N successful inserts,
// amortizing its cost instead of paying it on every request.
const cleanupEveryN = 32

// Store implements the Store interface with SQLite as backend.
//
// SQLite driver registration is the caller's responsibility — govisual does
// not import a driver to avoid forcing a specific implementation on users.
// Register your preferred driver (e.g. mattn/go-sqlite3 or ncruces/go-sqlite3)
// before calling NewStore, or use NewWithDB with a pre-built
// *sql.DB.
type Store struct {
	db             *sql.DB
	tableName      string
	capacity       int
	ownsConnection bool
	insertCount    atomic.Uint64
}

// NewStore creates a new SQLite-backed store.
// dbPath is forwarded to sql.Open("sqlite3", dbPath); ensure a SQLite driver
// is already registered under the name "sqlite3".
func New(dbPath, tableName string, capacity int) (*Store, error) {
	if capacity <= 0 {
		capacity = 100
	}

	if !store.IsValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name %q: must match [A-Za-z_][A-Za-z0-9_]*", tableName)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite DB: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQLite DB: %w", err)
	}

	s := &Store{
		db:             db,
		tableName:      tableName,
		capacity:       capacity,
		ownsConnection: true,
	}

	if err := s.createTable(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return s, nil
}

// NewWithDB creates a new SQLite store with an existing database connection.
func NewWithDB(db *sql.DB, tableName string, capacity int) (*Store, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}
	if capacity <= 0 {
		capacity = 100
	}
	if !store.IsValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name %q: must match [A-Za-z_][A-Za-z0-9_]*", tableName)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite DB: %w", err)
	}

	s := &Store{
		db:             db,
		tableName:      tableName,
		capacity:       capacity,
		ownsConnection: false,
	}

	if err := s.createTable(); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return s, nil
}

func (s *Store) createTable() error {
	// New tables carry an `extras` column for fields added in v2 (logs,
	// panic stack, performance metrics) as a single JSON blob. That keeps
	// the schema stable when we add more capture data later.
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			timestamp DATETIME,
			method TEXT,
			host TEXT,
			path TEXT,
			raw_path TEXT,
			query TEXT,
			request_headers TEXT,
			response_headers TEXT,
			status_code INTEGER,
			duration INTEGER,
			request_body TEXT,
			response_body TEXT,
			error TEXT,
			middleware_trace TEXT,
			route_trace TEXT,
			extras TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`, s.tableName)

	if _, err := s.db.Exec(query); err != nil {
		return err
	}

	for _, column := range []string{"extras", "host", "raw_path"} {
		if err := s.ensureTextColumn(column); err != nil {
			return err
		}
	}

	indexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s_timestamp_idx ON %s(timestamp DESC)",
		s.tableName, s.tableName)
	_, err := s.db.Exec(indexQuery)
	return err
}

// ensureTextColumn upgrades an existing table without relying on an ALTER
// error to determine whether the column already exists. The duplicate-column
// check only handles the race where another process migrates after the PRAGMA.
func (s *Store) ensureTextColumn(column string) error {
	rows, err := s.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", s.tableName))
	if err != nil {
		return fmt.Errorf("inspect schema for %s: %w", column, err)
	}

	found := false
	for rows.Next() {
		var (
			cid          int
			name         string
			columnType   string
			notNull      int
			defaultValue interface{}
			primaryKey   int
		)
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			rows.Close()
			return fmt.Errorf("inspect schema for %s: %w", column, err)
		}
		if strings.EqualFold(name, column) {
			found = true
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("inspect schema for %s: %w", column, err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("inspect schema for %s: %w", column, err)
	}
	if found {
		return nil
	}

	alter := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s TEXT", s.tableName, column)
	if _, err := s.db.Exec(alter); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return nil
		}
		return fmt.Errorf("add %s column: %w", column, err)
	}
	return nil
}

// extrasPayload holds the capture fields introduced in v2 that don't have
// their own columns. Serialized to the `extras` JSON column.
type extrasPayload struct {
	Logs               []store.LogEntry          `json:"logs,omitempty"`
	PanicStack         string                    `json:"panic_stack,omitempty"`
	PerformanceMetrics *store.PerformanceMetrics `json:"performance_metrics,omitempty"`
}

func encodeExtras(l *store.RequestLog) (string, error) {
	p := extrasPayload{
		Logs:               l.Logs,
		PanicStack:         l.PanicStack,
		PerformanceMetrics: l.PerformanceMetrics,
	}
	if len(p.Logs) == 0 && p.PanicStack == "" && p.PerformanceMetrics == nil {
		return "", nil
	}
	data, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("marshal extras: %w", err)
	}
	return string(data), nil
}

func decodeExtras(s string, l *store.RequestLog) {
	if s == "" {
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
		INSERT OR REPLACE INTO %s (
			id, timestamp, method, host, path, raw_path, query, request_headers, response_headers,
			status_code, duration, request_body, response_body, error,
			middleware_trace, route_trace, extras
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		return fmt.Errorf("sqlite insert: %w", err)
	}

	if s.insertCount.Add(1)%cleanupEveryN == 0 {
		if err := s.cleanup(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) cleanup() error {
	// One statement that keeps the newest rows; a separate COUNT would go
	// stale under concurrent inserts and leave the table above capacity.
	deleteQuery := fmt.Sprintf(`
		DELETE FROM %s
		WHERE id NOT IN (
			SELECT id FROM %s
			ORDER BY timestamp DESC, rowid DESC
			LIMIT ?
		)
	`, s.tableName, s.tableName)

	if _, err := s.db.Exec(deleteQuery, s.capacity); err != nil {
		return fmt.Errorf("sqlite cleanup: %w", err)
	}
	return nil
}

func (s *Store) Get(id string) (*store.RequestLog, bool) {
	query := fmt.Sprintf(`
		SELECT
			id, timestamp, method, COALESCE(host, ''), path, COALESCE(raw_path, ''), query,
			COALESCE(request_headers, '{}'),
			COALESCE(response_headers, '{}'),
			status_code, duration, request_body, response_body, error,
			COALESCE(middleware_trace, '[]'),
			COALESCE(route_trace, '{}'),
			COALESCE(extras, '')
		FROM %s
		WHERE id = ?
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
		log.Printf("govisual: failed to get request log from SQLite: %v", err)
		return nil, false
	}

	unmarshalLogJSON(reqHeadersStr, &reqLog.RequestHeaders, "request_headers", reqLog.ID)
	unmarshalLogJSON(respHeadersStr, &reqLog.ResponseHeaders, "response_headers", reqLog.ID)
	unmarshalLogJSON(middlewareTrace, &reqLog.MiddlewareTrace, "middleware_trace", reqLog.ID)
	unmarshalLogJSON(routeTrace, &reqLog.RouteTrace, "route_trace", reqLog.ID)
	decodeExtras(extras, &reqLog)

	return &reqLog, true
}

func (s *Store) GetAll() []*store.RequestLog {
	query := fmt.Sprintf(`
		SELECT
			id, timestamp, method, COALESCE(host, ''), path, COALESCE(raw_path, ''), query,
			COALESCE(request_headers, '{}'),
			COALESCE(response_headers, '{}'),
			status_code, duration, request_body, response_body, error,
			COALESCE(middleware_trace, '[]'),
			COALESCE(route_trace, '{}'),
			COALESCE(extras, '')
		FROM %s
		ORDER BY timestamp DESC
	`, s.tableName)

	return s.queryLogs(query)
}

func (s *Store) GetLatest(n int) []*store.RequestLog {
	query := fmt.Sprintf(`
		SELECT
			id, timestamp, method, COALESCE(host, ''), path, COALESCE(raw_path, ''), query,
			COALESCE(request_headers, '{}'),
			COALESCE(response_headers, '{}'),
			status_code, duration, request_body, response_body, error,
			COALESCE(middleware_trace, '[]'),
			COALESCE(route_trace, '{}'),
			COALESCE(extras, '')
		FROM %s
		ORDER BY timestamp DESC
		LIMIT ?
	`, s.tableName)

	return s.queryLogs(query, n)
}

func (s *Store) queryLogs(query string, args ...interface{}) []*store.RequestLog {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("govisual: failed to query logs from SQLite: %v", err)
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

func (s *Store) Clear() error {
	query := fmt.Sprintf("DELETE FROM %s", s.tableName)
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("failed to clear logs: %w", err)
	}
	return nil
}

func (s *Store) Close() error {
	if s.ownsConnection {
		return s.db.Close()
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

func unmarshalLogJSON(s string, v interface{}, field, logID string) {
	if s == "" {
		return
	}
	if err := json.Unmarshal([]byte(s), v); err != nil {
		log.Printf("govisual: failed to unmarshal %s for log %s: %v", field, logID, err)
	}
}

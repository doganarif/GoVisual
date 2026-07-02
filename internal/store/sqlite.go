package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/doganarif/govisual/internal/model"
)

// SQLiteStore implements the Store interface with SQLite as backend.
//
// SQLite driver registration is the caller's responsibility — govisual does
// not import a driver to avoid forcing a specific implementation on users.
// Register your preferred driver (e.g. mattn/go-sqlite3 or ncruces/go-sqlite3)
// before calling NewSQLiteStore, or use NewSQLiteStoreWithDB with a pre-built
// *sql.DB.
type SQLiteStore struct {
	db             *sql.DB
	tableName      string
	capacity       int
	ownsConnection bool
	insertCount    atomic.Uint64
}

// NewSQLiteStore creates a new SQLite-backed store.
// dbPath is forwarded to sql.Open("sqlite3", dbPath); ensure a SQLite driver
// is already registered under the name "sqlite3".
func NewSQLiteStore(dbPath, tableName string, capacity int) (*SQLiteStore, error) {
	if capacity <= 0 {
		capacity = 100
	}

	if !IsValidTableName(tableName) {
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

	s := &SQLiteStore{
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

// NewSQLiteStoreWithDB creates a new SQLite store with an existing database connection.
func NewSQLiteStoreWithDB(db *sql.DB, tableName string, capacity int) (*SQLiteStore, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}
	if capacity <= 0 {
		capacity = 100
	}
	if !IsValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name %q: must match [A-Za-z_][A-Za-z0-9_]*", tableName)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite DB: %w", err)
	}

	s := &SQLiteStore{
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

func (s *SQLiteStore) createTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			timestamp DATETIME,
			method TEXT,
			path TEXT,
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
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`, s.tableName)

	if _, err := s.db.Exec(query); err != nil {
		return err
	}

	indexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s_timestamp_idx ON %s(timestamp DESC)",
		s.tableName, s.tableName)
	_, err := s.db.Exec(indexQuery)
	return err
}

func (s *SQLiteStore) Add(reqLog *model.RequestLog) error {
	reqHeaders := prepareJSON(reqLog.RequestHeaders)
	respHeaders := prepareJSON(reqLog.ResponseHeaders)

	middlewareTrace := "[]"
	if len(reqLog.MiddlewareTrace) > 0 {
		if data, err := json.Marshal(reqLog.MiddlewareTrace); err == nil {
			middlewareTrace = string(data)
		}
	}

	routeTrace := "{}"
	if reqLog.RouteTrace != nil {
		if data, err := json.Marshal(reqLog.RouteTrace); err == nil {
			routeTrace = string(data)
		}
	}

	query := fmt.Sprintf(`
		INSERT OR REPLACE INTO %s (
			id, timestamp, method, path, query, request_headers, response_headers,
			status_code, duration, request_body, response_body, error,
			middleware_trace, route_trace
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, s.tableName)

	_, err := s.db.Exec(
		query,
		reqLog.ID,
		reqLog.Timestamp,
		reqLog.Method,
		reqLog.Path,
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
	)
	if err != nil {
		return fmt.Errorf("sqlite insert: %w", err)
	}

	if s.insertCount.Add(1)%cleanupEveryN == 0 {
		s.cleanup()
	}
	return nil
}

func (s *SQLiteStore) cleanup() {
	// One statement that keeps the newest rows; a separate COUNT would go
	// stale under concurrent inserts and leave the table above capacity.
	deleteQuery := fmt.Sprintf(`
		DELETE FROM %s
		WHERE id NOT IN (
			SELECT id FROM %s
			ORDER BY created_at DESC, timestamp DESC
			LIMIT ?
		)
	`, s.tableName, s.tableName)

	if _, err := s.db.Exec(deleteQuery, s.capacity); err != nil {
		log.Printf("govisual: failed to clean up old logs: %v", err)
	}
}

func (s *SQLiteStore) Get(id string) (*model.RequestLog, bool) {
	query := fmt.Sprintf(`
		SELECT
			id, timestamp, method, path, query,
			COALESCE(request_headers, '{}'),
			COALESCE(response_headers, '{}'),
			status_code, duration, request_body, response_body, error,
			COALESCE(middleware_trace, '[]'),
			COALESCE(route_trace, '{}')
		FROM %s
		WHERE id = ?
	`, s.tableName)

	var (
		reqLog          model.RequestLog
		reqHeadersStr   string
		respHeadersStr  string
		middlewareTrace string
		routeTrace      string
	)

	err := s.db.QueryRow(query, id).Scan(
		&reqLog.ID,
		&reqLog.Timestamp,
		&reqLog.Method,
		&reqLog.Path,
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

	return &reqLog, true
}

func (s *SQLiteStore) GetAll() []*model.RequestLog {
	query := fmt.Sprintf(`
		SELECT
			id, timestamp, method, path, query,
			COALESCE(request_headers, '{}'),
			COALESCE(response_headers, '{}'),
			status_code, duration, request_body, response_body, error,
			COALESCE(middleware_trace, '[]'),
			COALESCE(route_trace, '{}')
		FROM %s
		ORDER BY timestamp DESC
	`, s.tableName)

	return s.queryLogs(query)
}

func (s *SQLiteStore) GetLatest(n int) []*model.RequestLog {
	query := fmt.Sprintf(`
		SELECT
			id, timestamp, method, path, query,
			COALESCE(request_headers, '{}'),
			COALESCE(response_headers, '{}'),
			status_code, duration, request_body, response_body, error,
			COALESCE(middleware_trace, '[]'),
			COALESCE(route_trace, '{}')
		FROM %s
		ORDER BY timestamp DESC
		LIMIT ?
	`, s.tableName)

	return s.queryLogs(query, n)
}

func (s *SQLiteStore) queryLogs(query string, args ...interface{}) []*model.RequestLog {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("govisual: failed to query logs from SQLite: %v", err)
		return nil
	}
	defer rows.Close()

	var logs []*model.RequestLog
	for rows.Next() {
		var (
			reqLog          model.RequestLog
			reqHeadersStr   string
			respHeadersStr  string
			middlewareTrace string
			routeTrace      string
		)
		if err := rows.Scan(
			&reqLog.ID,
			&reqLog.Timestamp,
			&reqLog.Method,
			&reqLog.Path,
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
		); err != nil {
			log.Printf("govisual: failed to scan row: %v", err)
			continue
		}

		unmarshalLogJSON(reqHeadersStr, &reqLog.RequestHeaders, "request_headers", reqLog.ID)
		unmarshalLogJSON(respHeadersStr, &reqLog.ResponseHeaders, "response_headers", reqLog.ID)
		unmarshalLogJSON(middlewareTrace, &reqLog.MiddlewareTrace, "middleware_trace", reqLog.ID)
		unmarshalLogJSON(routeTrace, &reqLog.RouteTrace, "route_trace", reqLog.ID)

		logs = append(logs, &reqLog)
	}

	if err := rows.Err(); err != nil {
		log.Printf("govisual: error iterating over rows: %v", err)
	}

	return logs
}

func (s *SQLiteStore) Clear() error {
	query := fmt.Sprintf("DELETE FROM %s", s.tableName)
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("failed to clear logs: %w", err)
	}
	return nil
}

func (s *SQLiteStore) Close() error {
	if s.ownsConnection {
		return s.db.Close()
	}
	return nil
}

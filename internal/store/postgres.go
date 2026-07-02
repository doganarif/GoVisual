package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/doganarif/govisual/internal/model"
	_ "github.com/lib/pq"
)

// cleanupEveryN runs the capacity-trim query once every N successful inserts,
// instead of on every Add. Trading a slight overshoot of the configured capacity
// for far less load on the database.
const cleanupEveryN = 32

// PostgresStore implements the Store interface with PostgreSQL as backend
type PostgresStore struct {
	db          *sql.DB
	tableName   string
	capacity    int
	insertCount atomic.Uint64
}

// NewPostgresStore creates a new PostgreSQL-backed store
func NewPostgresStore(connStr, tableName string, capacity int) (*PostgresStore, error) {
	if capacity <= 0 {
		capacity = 100
	}

	if !IsValidTableName(tableName) {
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

	s := &PostgresStore{
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

func (s *PostgresStore) createTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			timestamp TIMESTAMP WITH TIME ZONE,
			method TEXT,
			path TEXT,
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
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`, s.tableName)

	if _, err := s.db.Exec(query); err != nil {
		return err
	}

	indexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s_timestamp_idx ON %s (timestamp DESC)",
		s.tableName, s.tableName)
	_, err := s.db.Exec(indexQuery)
	return err
}

// Add adds a new request log to the store
func (s *PostgresStore) Add(reqLog *model.RequestLog) error {
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
		INSERT INTO %s (
			id, timestamp, method, path, query, request_headers, response_headers,
			status_code, duration, request_body, response_body, error,
			middleware_trace, route_trace
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8, $9, $10, $11, $12, $13::jsonb, $14::jsonb)
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
		return fmt.Errorf("postgres insert: %w", err)
	}

	if s.insertCount.Add(1)%cleanupEveryN == 0 {
		s.cleanup()
	}
	return nil
}

// prepareJSON ensures we have a valid JSON string
func prepareJSON(v interface{}) string {
	if v == nil {
		return "{}"
	}
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("govisual: failed to marshal JSON: %v", err)
		return "{}"
	}
	return string(data)
}

// cleanup removes old logs to maintain the capacity limit
func (s *PostgresStore) cleanup() {
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
		log.Printf("govisual: failed to clean up old logs: %v", err)
	}
}

// Get retrieves a specific request log by its ID
func (s *PostgresStore) Get(id string) (*model.RequestLog, bool) {
	query := fmt.Sprintf(`
		SELECT
			id, timestamp, method, path, query,
			COALESCE(request_headers::text, '{}'),
			COALESCE(response_headers::text, '{}'),
			status_code, duration, request_body, response_body, error,
			COALESCE(middleware_trace::text, '[]'),
			COALESCE(route_trace::text, '{}')
		FROM %s
		WHERE id = $1
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
		log.Printf("govisual: failed to get request log from PostgreSQL: %v", err)
		return nil, false
	}

	unmarshalLogJSON(reqHeadersStr, &reqLog.RequestHeaders, "request_headers", reqLog.ID)
	unmarshalLogJSON(respHeadersStr, &reqLog.ResponseHeaders, "response_headers", reqLog.ID)
	unmarshalLogJSON(middlewareTrace, &reqLog.MiddlewareTrace, "middleware_trace", reqLog.ID)
	unmarshalLogJSON(routeTrace, &reqLog.RouteTrace, "route_trace", reqLog.ID)

	return &reqLog, true
}

// GetAll returns all stored request logs
func (s *PostgresStore) GetAll() []*model.RequestLog {
	query := fmt.Sprintf(`
		SELECT
			id, timestamp, method, path, query,
			COALESCE(request_headers::text, '{}'),
			COALESCE(response_headers::text, '{}'),
			status_code, duration, request_body, response_body, error,
			COALESCE(middleware_trace::text, '[]'),
			COALESCE(route_trace::text, '{}')
		FROM %s
		ORDER BY timestamp DESC
	`, s.tableName)

	return s.queryLogs(query)
}

// GetLatest returns the n most recent request logs
func (s *PostgresStore) GetLatest(n int) []*model.RequestLog {
	query := fmt.Sprintf(`
		SELECT
			id, timestamp, method, path, query,
			COALESCE(request_headers::text, '{}'),
			COALESCE(response_headers::text, '{}'),
			status_code, duration, request_body, response_body, error,
			COALESCE(middleware_trace::text, '[]'),
			COALESCE(route_trace::text, '{}')
		FROM %s
		ORDER BY timestamp DESC
		LIMIT $1
	`, s.tableName)

	return s.queryLogs(query, n)
}

func (s *PostgresStore) queryLogs(query string, args ...interface{}) []*model.RequestLog {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("govisual: failed to query logs from PostgreSQL: %v", err)
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

// Clear clears all stored request logs
func (s *PostgresStore) Clear() error {
	query := fmt.Sprintf("TRUNCATE TABLE %s", s.tableName)
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("failed to clear logs: %w", err)
	}
	return nil
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
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

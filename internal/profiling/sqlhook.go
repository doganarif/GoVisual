package profiling

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"
)

// SQLHook provides hooks for SQL operations
type SQLHook struct {
	profiler *Profiler
}

// NewSQLHook creates a new SQL hook
func NewSQLHook(profiler *Profiler) *SQLHook {
	return &SQLHook{profiler: profiler}
}

// WrapDriver wraps a SQL driver with profiling hooks
func (h *SQLHook) WrapDriver(d driver.Driver) driver.Driver {
	return &hookedDriver{driver: d, hook: h}
}

// hookedDriver wraps a driver.Driver with hooks
type hookedDriver struct {
	driver driver.Driver
	hook   *SQLHook
}

func (d *hookedDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.driver.Open(name)
	if err != nil {
		return nil, err
	}
	return &hookedConn{conn: conn, hook: d.hook}, nil
}

// hookedConn wraps a driver.Conn with hooks
type hookedConn struct {
	conn driver.Conn
	hook *SQLHook
}

func (c *hookedConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &hookedStmt{stmt: stmt, query: query, hook: c.hook}, nil
}

func (c *hookedConn) Close() error {
	return c.conn.Close()
}

func (c *hookedConn) Begin() (driver.Tx, error) {
	tx, err := c.conn.Begin()
	if err != nil {
		return nil, err
	}
	return &hookedTx{tx: tx, hook: c.hook}, nil
}

// Implement other required methods
func (c *hookedConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	var stmt driver.Stmt
	var err error

	start := time.Now()

	if prepCtx, ok := c.conn.(driver.ConnPrepareContext); ok {
		stmt, err = prepCtx.PrepareContext(ctx, query)
	} else {
		stmt, err = c.conn.Prepare(query)
	}

	duration := time.Since(start)
	c.hook.profiler.RecordSQLQuery(ctx, query, duration, 0, err)

	if err != nil {
		return nil, err
	}
	return &hookedStmt{stmt: stmt, query: query, hook: c.hook}, nil
}

func (c *hookedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	start := time.Now()

	var result driver.Result
	var err error

	if execCtx, ok := c.conn.(driver.ExecerContext); ok {
		result, err = execCtx.ExecContext(ctx, query, args)
	} else if exec, ok := c.conn.(driver.Execer); ok {
		values := make([]driver.Value, len(args))
		for i, arg := range args {
			values[i] = arg.Value
		}
		result, err = exec.Exec(query, values)
	} else {
		return nil, driver.ErrSkip
	}

	duration := time.Since(start)

	rows := int64(0)
	if result != nil {
		if r, err := result.RowsAffected(); err == nil {
			rows = r
		}
	}

	c.hook.profiler.RecordSQLQuery(ctx, query, duration, int(rows), err)

	return result, err
}

func (c *hookedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	start := time.Now()

	var rows driver.Rows
	var err error

	if queryCtx, ok := c.conn.(driver.QueryerContext); ok {
		rows, err = queryCtx.QueryContext(ctx, query, args)
	} else if queryer, ok := c.conn.(driver.Queryer); ok {
		values := make([]driver.Value, len(args))
		for i, arg := range args {
			values[i] = arg.Value
		}
		rows, err = queryer.Query(query, values)
	} else {
		return nil, driver.ErrSkip
	}

	duration := time.Since(start)
	c.hook.profiler.RecordSQLQuery(ctx, query, duration, 0, err)

	return rows, err
}

func (c *hookedConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	var tx driver.Tx
	var err error

	if beginTx, ok := c.conn.(driver.ConnBeginTx); ok {
		tx, err = beginTx.BeginTx(ctx, opts)
	} else {
		tx, err = c.conn.Begin()
	}

	if err != nil {
		return nil, err
	}
	return &hookedTx{tx: tx, hook: c.hook}, nil
}

// hookedStmt wraps a driver.Stmt with hooks
type hookedStmt struct {
	stmt  driver.Stmt
	query string
	hook  *SQLHook
}

func (s *hookedStmt) Close() error {
	return s.stmt.Close()
}

func (s *hookedStmt) NumInput() int {
	return s.stmt.NumInput()
}

func (s *hookedStmt) Exec(args []driver.Value) (driver.Result, error) {
	start := time.Now()
	result, err := s.stmt.Exec(args)
	duration := time.Since(start)

	rows := int64(0)
	if result != nil {
		if r, err := result.RowsAffected(); err == nil {
			rows = r
		}
	}

	// Use background context as we don't have access to request context here
	s.hook.profiler.RecordSQLQuery(context.Background(), s.query, duration, int(rows), err)

	return result, err
}

func (s *hookedStmt) Query(args []driver.Value) (driver.Rows, error) {
	start := time.Now()
	rows, err := s.stmt.Query(args)
	duration := time.Since(start)

	s.hook.profiler.RecordSQLQuery(context.Background(), s.query, duration, 0, err)

	return rows, err
}

func (s *hookedStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	start := time.Now()

	var result driver.Result
	var err error

	if stmtExecCtx, ok := s.stmt.(driver.StmtExecContext); ok {
		result, err = stmtExecCtx.ExecContext(ctx, args)
	} else {
		values := make([]driver.Value, len(args))
		for i, arg := range args {
			values[i] = arg.Value
		}
		result, err = s.stmt.Exec(values)
	}

	duration := time.Since(start)

	rows := int64(0)
	if result != nil {
		if r, err := result.RowsAffected(); err == nil {
			rows = r
		}
	}

	s.hook.profiler.RecordSQLQuery(ctx, s.query, duration, int(rows), err)

	return result, err
}

func (s *hookedStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	start := time.Now()

	var rows driver.Rows
	var err error

	if stmtQueryCtx, ok := s.stmt.(driver.StmtQueryContext); ok {
		rows, err = stmtQueryCtx.QueryContext(ctx, args)
	} else {
		values := make([]driver.Value, len(args))
		for i, arg := range args {
			values[i] = arg.Value
		}
		rows, err = s.stmt.Query(values)
	}

	duration := time.Since(start)
	s.hook.profiler.RecordSQLQuery(ctx, s.query, duration, 0, err)

	return rows, err
}

// hookedTx wraps a driver.Tx with hooks
type hookedTx struct {
	tx   driver.Tx
	hook *SQLHook
}

func (t *hookedTx) Commit() error {
	return t.tx.Commit()
}

func (t *hookedTx) Rollback() error {
	return t.tx.Rollback()
}

// RegisterSQLDriver registers a wrapped SQL driver with profiling support
func RegisterSQLDriver(name string, driver driver.Driver, profiler *Profiler) {
	hook := NewSQLHook(profiler)
	sql.Register(name+"-profiled", hook.WrapDriver(driver))
}

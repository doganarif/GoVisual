//lint:file-ignore SA1019 Legacy database/sql driver interfaces must remain forwarded for compatibility.

package profiling

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"time"
)

// WrapDriver wraps a database/sql driver so queries executed with a request
// context are recorded on that request's profile. Queries executed with
// contexts that carry no profile are passed through untouched.
func WrapDriver(d driver.Driver) driver.Driver {
	wrapped := &hookedDriver{driver: d}
	if driverContext, ok := d.(driver.DriverContext); ok {
		return &hookedDriverContext{
			hookedDriver:  wrapped,
			driverContext: driverContext,
		}
	}
	return wrapped
}

// hookedDriver wraps a driver.Driver with hooks
type hookedDriver struct {
	driver driver.Driver
}

func (d *hookedDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.driver.Open(name)
	if err != nil {
		return nil, err
	}
	return wrapConn(conn), nil
}

// hookedDriverContext keeps context-aware connection establishment available
// when the wrapped driver provides it. database/sql prefers OpenConnector over
// Driver.Open for these drivers.
type hookedDriverContext struct {
	*hookedDriver
	driverContext driver.DriverContext
}

func (d *hookedDriverContext) OpenConnector(name string) (driver.Connector, error) {
	connector, err := d.driverContext.OpenConnector(name)
	if err != nil {
		return nil, err
	}

	wrapped := &hookedConnector{connector: connector, driver: d}
	if closer, ok := connector.(io.Closer); ok {
		return &hookedConnectorCloser{hookedConnector: wrapped, closer: closer}, nil
	}
	return wrapped, nil
}

type hookedConnector struct {
	connector driver.Connector
	driver    driver.Driver
}

func (c *hookedConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return wrapConn(conn), nil
}

func (c *hookedConnector) Driver() driver.Driver {
	return c.driver
}

type hookedConnectorCloser struct {
	*hookedConnector
	closer io.Closer
}

func (c *hookedConnectorCloser) Close() error {
	return c.closer.Close()
}

// hookedConn wraps a driver.Conn with hooks
type hookedConn struct {
	conn driver.Conn
}

// hookedConnCore lists the context-aware interfaces implemented by hookedConn
// itself. Embedding this interface in the conditional wrappers below keeps
// query instrumentation active while exposing only the lifecycle and value
// conversion interfaces supported by the underlying connection.
type hookedConnCore interface {
	driver.Conn
	driver.ConnPrepareContext
	driver.ExecerContext
	driver.QueryerContext
	driver.ConnBeginTx
}

func wrapConn(conn driver.Conn) driver.Conn {
	core := hookedConnCore(&hookedConn{conn: conn})
	pinger, hasPinger := conn.(driver.Pinger)
	resetter, hasResetter := conn.(driver.SessionResetter)
	validator, hasValidator := conn.(driver.Validator)
	checker, hasChecker := conn.(driver.NamedValueChecker)

	mask := 0
	if hasPinger {
		mask |= 1
	}
	if hasResetter {
		mask |= 2
	}
	if hasValidator {
		mask |= 4
	}
	if hasChecker {
		mask |= 8
	}

	switch mask {
	case 1:
		return struct {
			hookedConnCore
			driver.Pinger
		}{core, pinger}
	case 2:
		return struct {
			hookedConnCore
			driver.SessionResetter
		}{core, resetter}
	case 3:
		return struct {
			hookedConnCore
			driver.Pinger
			driver.SessionResetter
		}{core, pinger, resetter}
	case 4:
		return struct {
			hookedConnCore
			driver.Validator
		}{core, validator}
	case 5:
		return struct {
			hookedConnCore
			driver.Pinger
			driver.Validator
		}{core, pinger, validator}
	case 6:
		return struct {
			hookedConnCore
			driver.SessionResetter
			driver.Validator
		}{core, resetter, validator}
	case 7:
		return struct {
			hookedConnCore
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{core, pinger, resetter, validator}
	case 8:
		return struct {
			hookedConnCore
			driver.NamedValueChecker
		}{core, checker}
	case 9:
		return struct {
			hookedConnCore
			driver.Pinger
			driver.NamedValueChecker
		}{core, pinger, checker}
	case 10:
		return struct {
			hookedConnCore
			driver.SessionResetter
			driver.NamedValueChecker
		}{core, resetter, checker}
	case 11:
		return struct {
			hookedConnCore
			driver.Pinger
			driver.SessionResetter
			driver.NamedValueChecker
		}{core, pinger, resetter, checker}
	case 12:
		return struct {
			hookedConnCore
			driver.Validator
			driver.NamedValueChecker
		}{core, validator, checker}
	case 13:
		return struct {
			hookedConnCore
			driver.Pinger
			driver.Validator
			driver.NamedValueChecker
		}{core, pinger, validator, checker}
	case 14:
		return struct {
			hookedConnCore
			driver.SessionResetter
			driver.Validator
			driver.NamedValueChecker
		}{core, resetter, validator, checker}
	case 15:
		return struct {
			hookedConnCore
			driver.Pinger
			driver.SessionResetter
			driver.Validator
			driver.NamedValueChecker
		}{core, pinger, resetter, validator, checker}
	default:
		return core
	}
}

func (c *hookedConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return wrapStmt(stmt, query), nil
}

func (c *hookedConn) Close() error {
	return c.conn.Close()
}

func (c *hookedConn) Begin() (driver.Tx, error) {
	tx, err := c.conn.Begin()
	if err != nil {
		return nil, err
	}
	return &hookedTx{tx: tx}, nil
}

// Implement other required methods
func (c *hookedConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	var stmt driver.Stmt
	var err error

	start := time.Now()

	if prepCtx, ok := c.conn.(driver.ConnPrepareContext); ok {
		stmt, err = prepCtx.PrepareContext(ctx, query)
	} else {
		if err := ctx.Err(); err != nil {
			RecordSQL(ctx, query, time.Since(start), 0, err)
			return nil, err
		}
		stmt, err = c.conn.Prepare(query)
		if err == nil {
			select {
			case <-ctx.Done():
				_ = stmt.Close()
				err = ctx.Err()
			default:
			}
		}
	}

	duration := time.Since(start)
	RecordSQL(ctx, query, duration, 0, err)

	if err != nil {
		return nil, err
	}
	return wrapStmt(stmt, query), nil
}

func (c *hookedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	start := time.Now()

	var result driver.Result
	var err error

	if execCtx, ok := c.conn.(driver.ExecerContext); ok {
		result, err = execCtx.ExecContext(ctx, query, args)
	} else if exec, ok := c.conn.(driver.Execer); ok {
		if err = ctx.Err(); err == nil {
			var values []driver.Value
			values, err = legacyValues(args)
			if err == nil {
				result, err = exec.Exec(query, values)
			}
		}
	} else {
		return nil, driver.ErrSkip
	}
	if errors.Is(err, driver.ErrSkip) {
		return result, err
	}

	duration := time.Since(start)

	rows := int64(0)
	if result != nil {
		if r, err := result.RowsAffected(); err == nil {
			rows = r
		}
	}

	RecordSQL(ctx, query, duration, int(rows), err)

	return result, err
}

func (c *hookedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	start := time.Now()

	var rows driver.Rows
	var err error

	if queryCtx, ok := c.conn.(driver.QueryerContext); ok {
		rows, err = queryCtx.QueryContext(ctx, query, args)
	} else if queryer, ok := c.conn.(driver.Queryer); ok {
		if err = ctx.Err(); err == nil {
			var values []driver.Value
			values, err = legacyValues(args)
			if err == nil {
				rows, err = queryer.Query(query, values)
			}
		}
	} else {
		return nil, driver.ErrSkip
	}
	if errors.Is(err, driver.ErrSkip) {
		return rows, err
	}

	duration := time.Since(start)
	RecordSQL(ctx, query, duration, 0, err)

	return rows, err
}

func (c *hookedConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	var tx driver.Tx
	var err error

	if beginTx, ok := c.conn.(driver.ConnBeginTx); ok {
		tx, err = beginTx.BeginTx(ctx, opts)
	} else {
		switch {
		case opts.Isolation != driver.IsolationLevel(0):
			err = fmt.Errorf("sql: driver does not support non-default isolation level")
		case opts.ReadOnly:
			err = fmt.Errorf("sql: driver does not support read-only transactions")
		case ctx.Err() != nil:
			err = ctx.Err()
		default:
			tx, err = c.conn.Begin()
			if err == nil {
				select {
				case <-ctx.Done():
					_ = tx.Rollback()
					err = ctx.Err()
				default:
				}
			}
		}
	}

	if err != nil {
		return nil, err
	}
	return &hookedTx{tx: tx}, nil
}

// hookedStmt wraps a driver.Stmt with hooks
type hookedStmt struct {
	stmt  driver.Stmt
	query string
}

type hookedStmtCore interface {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
}

// Give ColumnConverter an embedding name distinct from its method. Embedding
// driver.ColumnConverter directly would make the field name shadow the
// promoted ColumnConverter method.
type stmtColumnConverter interface {
	driver.ColumnConverter
}

func wrapStmt(stmt driver.Stmt, query string) driver.Stmt {
	core := hookedStmtCore(&hookedStmt{stmt: stmt, query: query})
	checker, hasChecker := stmt.(driver.NamedValueChecker)
	converter, hasConverter := stmt.(driver.ColumnConverter)

	switch {
	case hasChecker && hasConverter:
		return struct {
			hookedStmtCore
			driver.NamedValueChecker
			stmtColumnConverter
		}{core, checker, converter}
	case hasChecker:
		return struct {
			hookedStmtCore
			driver.NamedValueChecker
		}{core, checker}
	case hasConverter:
		return struct {
			hookedStmtCore
			stmtColumnConverter
		}{core, converter}
	default:
		return core
	}
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
	RecordSQL(context.Background(), s.query, duration, int(rows), err)

	return result, err
}

func (s *hookedStmt) Query(args []driver.Value) (driver.Rows, error) {
	start := time.Now()
	rows, err := s.stmt.Query(args)
	duration := time.Since(start)

	RecordSQL(context.Background(), s.query, duration, 0, err)

	return rows, err
}

func (s *hookedStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	start := time.Now()

	var result driver.Result
	var err error

	if stmtExecCtx, ok := s.stmt.(driver.StmtExecContext); ok {
		result, err = stmtExecCtx.ExecContext(ctx, args)
	} else {
		if err = ctx.Err(); err == nil {
			var values []driver.Value
			values, err = legacyValues(args)
			if err == nil {
				result, err = s.stmt.Exec(values)
			}
		}
	}

	duration := time.Since(start)

	rows := int64(0)
	if result != nil {
		if r, err := result.RowsAffected(); err == nil {
			rows = r
		}
	}

	RecordSQL(ctx, s.query, duration, int(rows), err)

	return result, err
}

func (s *hookedStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	start := time.Now()

	var rows driver.Rows
	var err error

	if stmtQueryCtx, ok := s.stmt.(driver.StmtQueryContext); ok {
		rows, err = stmtQueryCtx.QueryContext(ctx, args)
	} else {
		if err = ctx.Err(); err == nil {
			var values []driver.Value
			values, err = legacyValues(args)
			if err == nil {
				rows, err = s.stmt.Query(values)
			}
		}
	}

	duration := time.Since(start)
	RecordSQL(ctx, s.query, duration, 0, err)

	return rows, err
}

func legacyValues(args []driver.NamedValue) ([]driver.Value, error) {
	values := make([]driver.Value, len(args))
	for i, arg := range args {
		if arg.Name != "" {
			return nil, fmt.Errorf("sql: driver does not support the use of Named Parameters")
		}
		values[i] = arg.Value
	}
	return values, nil
}

// hookedTx wraps a driver.Tx with hooks
type hookedTx struct {
	tx driver.Tx
}

func (t *hookedTx) Commit() error {
	return t.tx.Commit()
}

func (t *hookedTx) Rollback() error {
	return t.tx.Rollback()
}

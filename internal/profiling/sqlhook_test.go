//lint:file-ignore SA1019 Tests assert compatibility with legacy database/sql driver interfaces.

package profiling

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"
	"time"
)

type fakeDriver struct{}

func (fakeDriver) Open(name string) (driver.Conn, error) { return &fakeConn{}, nil }

type fakeConn struct{}

func (*fakeConn) Prepare(query string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (*fakeConn) Close() error                              { return nil }
func (*fakeConn) Begin() (driver.Tx, error)                 { return nil, driver.ErrSkip }

func (*fakeConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(3), nil
}

func TestWrapDriverRecordsQueriesOnProfile(t *testing.T) {
	sql.Register("fake-viz", WrapDriver(fakeDriver{}))
	db, err := sql.Open("fake-viz", "")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	profiler := NewProfiler(10)
	profiler.SetProfileType(ProfileMemory)
	profiler.SetThreshold(0)
	ctx := profiler.StartProfiling(context.Background(), "req-sql")

	if _, err := db.ExecContext(ctx, "UPDATE things SET x = 1"); err != nil {
		t.Fatalf("exec: %v", err)
	}

	metrics := profiler.EndProfiling(ctx)
	if metrics == nil {
		t.Fatal("expected metrics")
	}
	if len(metrics.SQLQueries) != 1 {
		t.Fatalf("expected 1 recorded query, got %d", len(metrics.SQLQueries))
	}
	q := metrics.SQLQueries[0]
	if q.Query != "UPDATE things SET x = 1" || q.Rows != 3 {
		t.Fatalf("recorded query = %+v", q)
	}
}

func TestWrapDriverNoProfileIsNoOp(t *testing.T) {
	sql.Register("fake-viz-2", WrapDriver(fakeDriver{}))
	db, err := sql.Open("fake-viz-2", "")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	// A context without an active profile must pass through untouched.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, "SELECT 1"); err != nil {
		t.Fatalf("exec: %v", err)
	}
}

func TestRecordHTTPAttachesToProfile(t *testing.T) {
	profiler := NewProfiler(10)
	profiler.SetProfileType(ProfileMemory)
	profiler.SetThreshold(0)
	ctx := profiler.StartProfiling(context.Background(), "req-http")

	RecordHTTP(ctx, "GET", "https://api.example.com/users", 42*time.Millisecond, 200, 512)

	metrics := profiler.EndProfiling(ctx)
	if metrics == nil {
		t.Fatal("expected metrics")
	}
	if len(metrics.HTTPCalls) != 1 {
		t.Fatalf("expected 1 recorded call, got %d", len(metrics.HTTPCalls))
	}
	c := metrics.HTTPCalls[0]
	if c.URL != "https://api.example.com/users" || c.Status != 200 || c.Size != 512 {
		t.Fatalf("recorded call = %+v", c)
	}
}

type optionalDriver struct {
	conn          *optionalConn
	connectorName string
	connectValue  string
	closed        bool
}

func (d *optionalDriver) Open(string) (driver.Conn, error) { return d.conn, nil }

func (d *optionalDriver) OpenConnector(name string) (driver.Connector, error) {
	d.connectorName = name
	return &optionalConnector{driver: d}, nil
}

type optionalConnector struct {
	driver *optionalDriver
}

func (c *optionalConnector) Connect(ctx context.Context) (driver.Conn, error) {
	c.driver.connectValue, _ = ctx.Value(optionalContextKey{}).(string)
	return c.driver.conn, nil
}

func (c *optionalConnector) Driver() driver.Driver { return c.driver }

func (c *optionalConnector) Close() error {
	c.driver.closed = true
	return nil
}

type optionalContextKey struct{}

type optionalConn struct {
	stmt       *optionalStmt
	pinged     bool
	reset      bool
	checked    bool
	valid      bool
	preparedBy string
}

func (c *optionalConn) Prepare(string) (driver.Stmt, error) {
	c.preparedBy = "legacy"
	return c.stmt, nil
}

func (c *optionalConn) PrepareContext(context.Context, string) (driver.Stmt, error) {
	c.preparedBy = "context"
	return c.stmt, nil
}

func (*optionalConn) Close() error              { return nil }
func (*optionalConn) Begin() (driver.Tx, error) { return optionalTx{}, nil }

func (c *optionalConn) Ping(context.Context) error {
	c.pinged = true
	return nil
}

func (c *optionalConn) ResetSession(context.Context) error {
	c.reset = true
	return nil
}

func (c *optionalConn) IsValid() bool { return c.valid }

func (c *optionalConn) CheckNamedValue(value *driver.NamedValue) error {
	c.checked = true
	value.Value = "conn-checked"
	return nil
}

type optionalStmt struct {
	execContextCalled  bool
	queryContextCalled bool
	checked            bool
	convertedIndex     int
}

func (*optionalStmt) Close() error  { return nil }
func (*optionalStmt) NumInput() int { return 1 }
func (*optionalStmt) Exec([]driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}
func (*optionalStmt) Query([]driver.Value) (driver.Rows, error) { return nil, nil }

func (s *optionalStmt) ExecContext(context.Context, []driver.NamedValue) (driver.Result, error) {
	s.execContextCalled = true
	return driver.RowsAffected(2), nil
}

func (s *optionalStmt) QueryContext(context.Context, []driver.NamedValue) (driver.Rows, error) {
	s.queryContextCalled = true
	return nil, nil
}

func (s *optionalStmt) CheckNamedValue(value *driver.NamedValue) error {
	s.checked = true
	value.Value = "stmt-checked"
	return nil
}

func (s *optionalStmt) ColumnConverter(index int) driver.ValueConverter {
	s.convertedIndex = index
	return driver.String
}

type optionalTx struct{}

func (optionalTx) Commit() error   { return nil }
func (optionalTx) Rollback() error { return nil }

func TestWrapDriverPreservesOptionalInterfaces(t *testing.T) {
	stmt := &optionalStmt{}
	conn := &optionalConn{stmt: stmt, valid: false}
	original := &optionalDriver{conn: conn}
	wrapped := WrapDriver(original)

	driverContext, ok := wrapped.(driver.DriverContext)
	if !ok {
		t.Fatal("wrapped driver dropped DriverContext")
	}
	connector, err := driverContext.OpenConnector("test-dsn")
	if err != nil {
		t.Fatalf("OpenConnector: %v", err)
	}
	if original.connectorName != "test-dsn" {
		t.Fatalf("connector name = %q", original.connectorName)
	}
	if connector.Driver() != wrapped {
		t.Fatal("wrapped connector returned the uninstrumented driver")
	}

	ctx := context.WithValue(context.Background(), optionalContextKey{}, "connect-context")
	wrappedConn, err := connector.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if original.connectValue != "connect-context" {
		t.Fatalf("Connect context value = %q", original.connectValue)
	}

	pinger, ok := wrappedConn.(driver.Pinger)
	if !ok {
		t.Fatal("wrapped connection dropped Pinger")
	}
	if err := pinger.Ping(ctx); err != nil || !conn.pinged {
		t.Fatalf("Ping was not forwarded: err=%v pinged=%v", err, conn.pinged)
	}

	resetter, ok := wrappedConn.(driver.SessionResetter)
	if !ok {
		t.Fatal("wrapped connection dropped SessionResetter")
	}
	if err := resetter.ResetSession(ctx); err != nil || !conn.reset {
		t.Fatalf("ResetSession was not forwarded: err=%v reset=%v", err, conn.reset)
	}

	validator, ok := wrappedConn.(driver.Validator)
	if !ok {
		t.Fatal("wrapped connection dropped Validator")
	}
	if validator.IsValid() {
		t.Fatal("IsValid did not return the underlying result")
	}

	checker, ok := wrappedConn.(driver.NamedValueChecker)
	if !ok {
		t.Fatal("wrapped connection dropped NamedValueChecker")
	}
	named := &driver.NamedValue{Value: "original"}
	if err := checker.CheckNamedValue(named); err != nil || !conn.checked || named.Value != "conn-checked" {
		t.Fatalf("connection CheckNamedValue was not forwarded: value=%v err=%v", named.Value, err)
	}

	prepareContext := wrappedConn.(driver.ConnPrepareContext)
	wrappedStmt, err := prepareContext.PrepareContext(ctx, "SELECT ?")
	if err != nil {
		t.Fatalf("PrepareContext: %v", err)
	}
	if conn.preparedBy != "context" {
		t.Fatalf("prepared through %q path", conn.preparedBy)
	}

	stmtChecker, ok := wrappedStmt.(driver.NamedValueChecker)
	if !ok {
		t.Fatal("wrapped statement dropped NamedValueChecker")
	}
	named.Value = "original"
	if err := stmtChecker.CheckNamedValue(named); err != nil || !stmt.checked || named.Value != "stmt-checked" {
		t.Fatalf("statement CheckNamedValue was not forwarded: value=%v err=%v", named.Value, err)
	}

	converter, ok := wrappedStmt.(driver.ColumnConverter)
	if !ok {
		t.Fatal("wrapped statement dropped ColumnConverter")
	}
	converted, err := converter.ColumnConverter(3).ConvertValue(42)
	if err != nil || converted != "42" || stmt.convertedIndex != 3 {
		t.Fatalf("ColumnConverter was not forwarded: value=%v index=%d err=%v", converted, stmt.convertedIndex, err)
	}

	if _, err := wrappedStmt.(driver.StmtExecContext).ExecContext(ctx, nil); err != nil || !stmt.execContextCalled {
		t.Fatalf("StmtExecContext was not forwarded: err=%v called=%v", err, stmt.execContextCalled)
	}
	if _, err := wrappedStmt.(driver.StmtQueryContext).QueryContext(ctx, nil); err != nil || !stmt.queryContextCalled {
		t.Fatalf("StmtQueryContext was not forwarded: err=%v called=%v", err, stmt.queryContextCalled)
	}

	closer, ok := connector.(io.Closer)
	if !ok {
		t.Fatal("wrapped connector dropped io.Closer")
	}
	if err := closer.Close(); err != nil || !original.closed {
		t.Fatalf("connector Close was not forwarded: err=%v closed=%v", err, original.closed)
	}
}

type minimalDriver struct{}

func (minimalDriver) Open(string) (driver.Conn, error) { return minimalConn{}, nil }

type minimalConn struct{}

func (minimalConn) Prepare(string) (driver.Stmt, error) { return minimalStmt{}, nil }
func (minimalConn) Close() error                        { return nil }
func (minimalConn) Begin() (driver.Tx, error)           { return optionalTx{}, nil }

type minimalStmt struct{}

func (minimalStmt) Close() error                               { return nil }
func (minimalStmt) NumInput() int                              { return -1 }
func (minimalStmt) Exec([]driver.Value) (driver.Result, error) { return driver.RowsAffected(0), nil }
func (minimalStmt) Query([]driver.Value) (driver.Rows, error)  { return nil, nil }

type legacyConn struct {
	prepareCalls  int
	execCalls     int
	queryCalls    int
	beginCalls    int
	prepareCancel context.CancelFunc
	beginCancel   context.CancelFunc
	stmt          *legacyStmt
	tx            *trackingTx
}

func (c *legacyConn) Prepare(string) (driver.Stmt, error) {
	c.prepareCalls++
	if c.prepareCancel != nil {
		c.prepareCancel()
	}
	if c.stmt == nil {
		c.stmt = &legacyStmt{}
	}
	return c.stmt, nil
}
func (*legacyConn) Close() error { return nil }
func (c *legacyConn) Begin() (driver.Tx, error) {
	c.beginCalls++
	if c.beginCancel != nil {
		c.beginCancel()
	}
	if c.tx == nil {
		c.tx = &trackingTx{}
	}
	return c.tx, nil
}
func (c *legacyConn) Exec(string, []driver.Value) (driver.Result, error) {
	c.execCalls++
	return driver.RowsAffected(1), nil
}
func (c *legacyConn) Query(string, []driver.Value) (driver.Rows, error) {
	c.queryCalls++
	return nil, nil
}

type legacyStmt struct {
	execCalls  int
	queryCalls int
	closed     bool
}

func (s *legacyStmt) Close() error { s.closed = true; return nil }
func (*legacyStmt) NumInput() int  { return -1 }
func (s *legacyStmt) Exec([]driver.Value) (driver.Result, error) {
	s.execCalls++
	return driver.RowsAffected(1), nil
}
func (s *legacyStmt) Query([]driver.Value) (driver.Rows, error) {
	s.queryCalls++
	return nil, nil
}

type trackingTx struct{ rolledBack bool }

func (*trackingTx) Commit() error { return nil }
func (t *trackingTx) Rollback() error {
	t.rolledBack = true
	return nil
}

func TestLegacyFallbacksMatchDatabaseSQLValidation(t *testing.T) {
	underlying := &legacyConn{}
	wrapped := &hookedConn{conn: underlying}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := wrapped.PrepareContext(canceled, "SELECT 1"); err != context.Canceled || underlying.prepareCalls != 0 {
		t.Fatalf("canceled PrepareContext: err=%v calls=%d", err, underlying.prepareCalls)
	}
	if _, err := wrapped.ExecContext(canceled, "UPDATE t", nil); err != context.Canceled || underlying.execCalls != 0 {
		t.Fatalf("canceled ExecContext: err=%v calls=%d", err, underlying.execCalls)
	}
	if _, err := wrapped.QueryContext(canceled, "SELECT 1", nil); err != context.Canceled || underlying.queryCalls != 0 {
		t.Fatalf("canceled QueryContext: err=%v calls=%d", err, underlying.queryCalls)
	}
	if _, err := wrapped.ExecContext(context.Background(), "UPDATE t", []driver.NamedValue{{Name: "named", Value: 1}}); err == nil || underlying.execCalls != 0 {
		t.Fatalf("named ExecContext: err=%v calls=%d", err, underlying.execCalls)
	}
	if _, err := wrapped.QueryContext(context.Background(), "SELECT 1", []driver.NamedValue{{Name: "named", Value: 1}}); err == nil || underlying.queryCalls != 0 {
		t.Fatalf("named QueryContext: err=%v calls=%d", err, underlying.queryCalls)
	}

	for _, opts := range []driver.TxOptions{
		{Isolation: driver.IsolationLevel(1)},
		{ReadOnly: true},
	} {
		if _, err := wrapped.BeginTx(context.Background(), opts); err == nil {
			t.Fatalf("unsupported transaction options %+v were accepted", opts)
		}
	}
	if _, err := wrapped.BeginTx(canceled, driver.TxOptions{}); err != context.Canceled {
		t.Fatalf("canceled BeginTx error = %v", err)
	}
	if underlying.beginCalls != 0 {
		t.Fatalf("legacy Begin called %d times for rejected transactions", underlying.beginCalls)
	}
	if _, err := wrapped.BeginTx(context.Background(), driver.TxOptions{}); err != nil || underlying.beginCalls != 1 {
		t.Fatalf("default BeginTx: err=%v calls=%d", err, underlying.beginCalls)
	}

	prepareCtx, cancelPrepare := context.WithCancel(context.Background())
	prepareStmt := &legacyStmt{}
	prepareConn := &legacyConn{prepareCancel: cancelPrepare, stmt: prepareStmt}
	if _, err := (&hookedConn{conn: prepareConn}).PrepareContext(prepareCtx, "SELECT 1"); err != context.Canceled || !prepareStmt.closed {
		t.Fatalf("cancellation during legacy Prepare: err=%v closed=%v", err, prepareStmt.closed)
	}

	beginCtx, cancelBegin := context.WithCancel(context.Background())
	beginTx := &trackingTx{}
	beginConn := &legacyConn{beginCancel: cancelBegin, tx: beginTx}
	if _, err := (&hookedConn{conn: beginConn}).BeginTx(beginCtx, driver.TxOptions{}); err != context.Canceled || !beginTx.rolledBack {
		t.Fatalf("cancellation during legacy Begin: err=%v rolledBack=%v", err, beginTx.rolledBack)
	}

	stmt := &legacyStmt{}
	wrappedStmt := &hookedStmt{stmt: stmt, query: "SELECT ?"}
	if _, err := wrappedStmt.ExecContext(canceled, nil); err != context.Canceled || stmt.execCalls != 0 {
		t.Fatalf("canceled StmtExecContext: err=%v calls=%d", err, stmt.execCalls)
	}
	if _, err := wrappedStmt.QueryContext(canceled, nil); err != context.Canceled || stmt.queryCalls != 0 {
		t.Fatalf("canceled StmtQueryContext: err=%v calls=%d", err, stmt.queryCalls)
	}
	if _, err := wrappedStmt.ExecContext(context.Background(), []driver.NamedValue{{Name: "named", Value: 1}}); err == nil || stmt.execCalls != 0 {
		t.Fatalf("named StmtExecContext: err=%v calls=%d", err, stmt.execCalls)
	}
	if _, err := wrappedStmt.QueryContext(context.Background(), []driver.NamedValue{{Name: "named", Value: 1}}); err == nil || stmt.queryCalls != 0 {
		t.Fatalf("named StmtQueryContext: err=%v calls=%d", err, stmt.queryCalls)
	}
}

func TestWrapDriverDoesNotInventOptionalInterfaces(t *testing.T) {
	wrapped := WrapDriver(minimalDriver{})
	if _, ok := wrapped.(driver.DriverContext); ok {
		t.Fatal("wrapped driver unexpectedly implements DriverContext")
	}

	conn, err := wrapped.Open("")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, ok := conn.(driver.Pinger); ok {
		t.Fatal("wrapped connection unexpectedly implements Pinger")
	}
	if _, ok := conn.(driver.SessionResetter); ok {
		t.Fatal("wrapped connection unexpectedly implements SessionResetter")
	}
	if _, ok := conn.(driver.Validator); ok {
		t.Fatal("wrapped connection unexpectedly implements Validator")
	}
	if _, ok := conn.(driver.NamedValueChecker); ok {
		t.Fatal("wrapped connection unexpectedly implements NamedValueChecker")
	}

	stmt, err := conn.Prepare("SELECT ?")
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if _, ok := stmt.(driver.NamedValueChecker); ok {
		t.Fatal("wrapped statement unexpectedly implements NamedValueChecker")
	}
	if _, ok := stmt.(driver.ColumnConverter); ok {
		t.Fatal("wrapped statement unexpectedly implements ColumnConverter")
	}
	if _, ok := stmt.(driver.StmtExecContext); !ok {
		t.Fatal("wrapped statement dropped context-aware execution")
	}
	if _, ok := stmt.(driver.StmtQueryContext); !ok {
		t.Fatal("wrapped statement dropped context-aware querying")
	}
}

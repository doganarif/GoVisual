package profiling

import (
	"context"
	"database/sql"
	"database/sql/driver"
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

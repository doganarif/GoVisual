package govisual

import (
	"database/sql/driver"

	"github.com/doganarif/govisual/v2/internal/profiling"
)

// WrapDriver instruments a database/sql driver so queries executed with a
// request's context show up on that request's profile in the dashboard.
// Register the wrapped driver once and open the database through it:
//
//	sql.Register("postgres+viz", govisual.WrapDriver(&pq.Driver{}))
//	db, err := sql.Open("postgres+viz", dsn)
//
// Queries are attributed through the context, so they must run through the
// *Context variants (QueryContext, ExecContext) with the incoming request's
// context, and profiling must be enabled via WithProfiling(true).
func WrapDriver(d driver.Driver) driver.Driver {
	return profiling.WrapDriver(d)
}

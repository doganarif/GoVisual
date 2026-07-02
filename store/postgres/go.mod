module github.com/doganarif/govisual/store/postgres

go 1.24.0

require (
	github.com/doganarif/govisual/v2 v2.0.0
	github.com/lib/pq v1.12.3
)

replace github.com/doganarif/govisual/v2 => ../..

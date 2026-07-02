module github.com/doganarif/govisual/store/redis

go 1.24.0

require (
	github.com/doganarif/govisual/v2 v2.0.0
	github.com/go-redis/redis/v8 v8.11.5
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)

replace github.com/doganarif/govisual/v2 => ../..

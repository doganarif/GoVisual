module github.com/doganarif/govisual/v2/examples

go 1.25.0

require (
	github.com/doganarif/govisual/store/mongodb v0.0.0-00010101000000-000000000000
	github.com/doganarif/govisual/store/postgres v0.0.0-00010101000000-000000000000
	github.com/doganarif/govisual/store/redis v0.0.0-00010101000000-000000000000
	github.com/doganarif/govisual/store/sqlite v0.0.0-00010101000000-000000000000
	github.com/doganarif/govisual/telemetry v0.0.0-00010101000000-000000000000
	github.com/doganarif/govisual/v2 v2.0.0
	github.com/mattn/go-sqlite3 v1.14.47
	go.opentelemetry.io/otel v1.44.0
	go.opentelemetry.io/otel/trace v1.44.0
)

require (
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/google/pprof v0.0.0-20240227163752-401108e1b7e7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/lib/pq v1.12.3 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.mongodb.org/mongo-driver/v2 v2.7.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk v1.44.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	golang.org/x/crypto v0.51.0 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/grpc v1.81.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	github.com/doganarif/govisual/store/mongodb => ../../store/mongodb
	github.com/doganarif/govisual/store/postgres => ../../store/postgres
	github.com/doganarif/govisual/store/redis => ../../store/redis
	github.com/doganarif/govisual/store/sqlite => ../../store/sqlite
	github.com/doganarif/govisual/telemetry => ../../telemetry
	github.com/doganarif/govisual/v2 => ../..
)

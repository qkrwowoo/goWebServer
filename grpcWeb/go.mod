module grpcWeb

go 1.23.0

toolchain go1.23.10

require (
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/pborman/getopt v1.1.0
	google.golang.org/grpc v1.73.0
)

require golang.org/x/text v0.26.0 // indirect

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/VictoriaMetrics/easyproto v0.1.4 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/godror/knownpb v0.3.0 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/suapapa/go_hangul v1.2.1 // indirect
	golang.org/x/exp v0.0.0-20250506013437-ce4c2cf36ca6 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/denisenkom/go-mssqldb v0.12.3
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-sql-driver/mysql v1.9.3
	github.com/godror/godror v0.48.3
	golang.org/x/crypto v0.39.0
	google.golang.org/protobuf v1.36.6
	local/common v0.0.0
)

replace local/common => ../common

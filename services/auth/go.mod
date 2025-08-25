module github.com/VariableSan/go-factory-microservice/services/auth

go 1.25.0

require (
	github.com/VariableSan/go-factory-microservice/pkg/common v0.0.0
	github.com/VariableSan/go-factory-microservice/pkg/proto v0.0.0
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.1
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	golang.org/x/crypto v0.39.0
	google.golang.org/grpc v1.74.2
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/lib/pq v1.10.9 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/protobuf v1.36.7 // indirect
)

replace github.com/VariableSan/go-factory-microservice/pkg/common => ../../pkg/common

replace github.com/VariableSan/go-factory-microservice/pkg/proto => ../../pkg/proto

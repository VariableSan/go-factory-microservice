module github.com/VariableSan/go-factory-microservice/services/feed

go 1.25.0

replace github.com/VariableSan/go-factory-microservice/pkg/common => ../../pkg/common

replace github.com/VariableSan/go-factory-microservice/pkg/proto => ../../pkg/proto

require (
	github.com/VariableSan/go-factory-microservice/pkg/common v0.0.0-00010101000000-000000000000
	github.com/golang-migrate/migrate/v4 v4.18.1
)

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	go.uber.org/atomic v1.7.0 // indirect
)

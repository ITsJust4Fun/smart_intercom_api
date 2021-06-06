## Generate proto
protoc --go_out=proto --go_opt=paths=source_relative --go-grpc_out=proto --go-grpc_opt=paths=source_relative proto/diagnostics.proto

## Generate GraphQL
go run github.com/99designs/gqlgen generate

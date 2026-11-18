// Package tools also hosts the protobuf codegen directive. Regenerate with:
//
//	go generate ./...
package tools

//go:generate protoc -Iproto -I/usr/include --go_out=internal/gen --go_opt=paths=source_relative --go-grpc_out=internal/gen --go-grpc_opt=paths=source_relative proto/trustbox/v1/control_plane.proto proto/trustbox/v1/node_daemon.proto

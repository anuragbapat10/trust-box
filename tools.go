//go:build tools

// Package tools pins protoc plugin versions. Install with:
// go install google.golang.org/protobuf/cmd/protoc-gen-go \
//   google.golang.org/grpc/cmd/protoc-gen-go-grpc
package tools

import (
	_ "google.golang.org/grpc" // runtime for generated *_grpc.pb.go code
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)

module github.com/menta2k/protoc-gen-redact/v3/examples/user

go 1.22.2

require (
	github.com/menta2k/protoc-gen-redact/v3 v3.0.0
	github.com/golang/protobuf v1.5.4
	google.golang.org/grpc v1.63.2
	google.golang.org/protobuf v1.33.0
)

require (
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240412170617-26222e5d3d56 // indirect
)

replace github.com/menta2k/protoc-gen-redact/v3 => ../..

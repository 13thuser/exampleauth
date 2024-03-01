# ExampleAuth

Example project to demonstrate auth using JWT token, and permissions access via claims


## Setup

1. Install `protoc` [https://grpc.io/docs/protoc-installation/]

2. Install the protocol compiler plugins for Go using the following commands:

    `$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28`

    `$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2`

3. Update your PATH so that the protoc compiler can find the plugins:

    `$ export PATH="$PATH:$(go env GOPATH)/bin"`

4. Generate grpc code:

    `make protos`

5. Run the unit tests:

    `make test` or `make test-verbose`

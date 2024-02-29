PROTO_DIR := protos
GO_OUT := grpc

.PHONY: all
all: protos

.PHONY: protos
protos:
	@echo "Generating Go protobuf files..."
	protoc -I=$(PROTO_DIR) --go_out=$(GO_OUT) --go_opt=paths=source_relative --go-grpc_out=$(GO_OUT)  --go-grpc_opt=paths=source_relative $(PROTO_DIR)/*.proto

.PHONY: clean
clean:
	@echo "Cleaning generated files..."
	rm -rf $(GO_OUT)/*.go

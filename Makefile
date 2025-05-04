.PHONY: protogen
protogen:
	protoc ./subpub.proto --go_out=./protogen --go_opt=paths=source_relative --go-grpc_out=./protogen --go-grpc_opt=paths=source_relative

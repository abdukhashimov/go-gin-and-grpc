CURRENT_DIR=$(shell pwd)

proto-gen:
	./scripts/gen-proto.sh	${CURRENT_DIR}

pull-proto-module:
	git submodule update --init --recursive

update-proto-module:
	git submodule update --remote --merge && rm -rf protos/* && cp -R go-gin-grpc-protos/* protos

swag_init:
	swag init -g api/main.go -o api/docs

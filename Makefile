CURRENT_DIR=$(shell pwd)

proto-gen:
	./scripts/gen-proto.sh	${CURRENT_DIR}

pull-proto-module:
	git submodule update --init --recursive

update-proto-module:
	git submodule update --remote --merge

swag_init:
	swag init -g api/main.go -o api/docs

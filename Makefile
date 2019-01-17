.PHONY:all

TMP_GO_PATH := .tmp_go_path
OUT_DIR := out
BIN_NAME := acc-k8s

TMP_GO_ABS := $(CURDIR)/$(TMP_GO_PATH)
ENV_GOPATH := env GOPATH=$(TMP_GO_ABS)

SOURCES := $(shell find . 2>&1 | grep -E '.*\.(c|h|go)$$')

all: $(SOURCES)
	go get -u github.com/golang/protobuf/protoc-gen-go
	protoc --go_out=. ./device_proto/device.proto
	$(ENV_GOPATH) go get -d -v ./...
	find . -path '*/vendor' | xargs -IX rm -rf X
	$(ENV_GOPATH) go build -ldflags "-s -w -v" -o $(OUT_DIR)/$(BIN_NAME) .

install:
	\cp $(CURDIR)/$(OUT_DIR)/$(BIN_NAME) /usr/bin/$(BIN_NAME)

clean:
	rm -rf $(OUT_DIR)
	rm -rf $(TMP_GO_PATH)
	if [ -f "./device_proto/device.pb.go" ]; then rm ./device_proto/device.pb.go; fi


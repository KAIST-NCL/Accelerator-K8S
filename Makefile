.PHONY:all


OUT_DIR := out
BIN_NAME := fpga-k8s

TMP_GO_ABS := $(CURDIR)/vendor
ENV_GOPATH := env GOPATH=$(TMP_GO_ABS)

SOURCES := $(shell find . 2>&1 | grep -E '.*\.(c|h|go)$$')

all: $(SOURCES)
	@if [ ! -d "./$(OUT_DIR)" ]; then mkdir $(OUT_DIR); fi
	protoc --go_out=. ./device_proto/device.proto
	$(ENV_GOPATH) go build -ldflags "-s -w -v" -o $(OUT_DIR)/$(BIN_NAME) .

install:
	\cp $(CURDIR)/$(OUT_DIR)/$(BIN_NAME) /usr/bin/$(BIN_NAME)

clean:
	rm -rf $(OUT_DIR)
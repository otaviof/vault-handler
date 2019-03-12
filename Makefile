APP = vault-handler
BUILD_DIR ?= build

.PHONY: default bootstrap build clean test

default: build

bootstrap:
	dep ensure -v -vendor-only

build: clean
	go build -v -o $(BUILD_DIR)/$(APP) cmd/$(APP)/*

clean:
	rm -rf $(BUILD_DIR) > /dev/null

clean-vendor:
	rm -rf ./vendor > /dev/null

test:
	go test -cover -v pkg/$(APP)/*
# application name
APP = vault-handler
# build directory
BUILD_DIR ?= build
# docker image tag
DOCKER_IMAGE ?= "otaviof/$(APP)"
# directory containing end-to-end tests
E2E_TEST_DIR ?= test/e2e
# project version, used as docker tag
VERSION ?= $(shell cat ./version)

.PHONY: default bootstrap build clean test

default: build

dep:
	go get -u github.com/golang/dep/cmd/dep

bootstrap:
	dep ensure -v -vendor-only

build: clean
	go build -v -o $(BUILD_DIR)/$(APP) cmd/$(APP)/*

build-docker:
	docker build --tag $(DOCKER_IMAGE):$(VERSION) .

clean:
	rm -rf $(BUILD_DIR) > /dev/null

clean-vendor:
	rm -rf ./vendor > /dev/null

run:
	go run -v cmd/$(APP)/* $(filter-out $@,$(MAKECMDGOALS))

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -cover -v pkg/$(APP)/*

snapshot:
	goreleaser --rm-dist --snapshot

release:
	git tag $(VERSION)
	git push origin $(VERSION)
	goreleaser --rm-dist

integration:
	go test -v $(E2E_TEST_DIR)/*

codecov:
	mkdir .ci || true
	curl -s -o .ci/codecov.sh https://codecov.io/bash
	bash .ci/codecov.sh -t $(CODECOV_TOKEN)

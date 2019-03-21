APP = vault-handler
BUILD_DIR ?= build
DOCKER_IMAGE ?= "otaviof/$(APP)"
VERSION ?= $(shell cat ./version)

.PHONY: default bootstrap build clean test

default: build

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

test:
	go test -cover -v pkg/$(APP)/*

snapshot:
	goreleaser --rm-dist --snapshot

release: release-go release-docker
	@echo "# Uploaded vault-handler v'$(VERSION)'!"

release-go:
	git tag --annotate $(VERSION)
	git push origin $(VERSION)
	goreleaser --rm-dist

release-docker: build-docker
	docker push $(DOCKER_IMAGE):$(VERSION)

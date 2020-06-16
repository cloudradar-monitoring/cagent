.PHONY: all

PROJECT_DIR=/go/src/github.com/cloudradar-monitoring/cagent

ifeq ($(RELEASE_MODE),)
  RELEASE_MODE=release-candidate
endif
ifeq ($(RELEASE_MODE),release-candidate)
  SELF_UPDATES_FEED_URL="https://repo.cloudradar.io/windows/cagent/feed/rolling"
  PROPRIETARY_SELF_UPDATES_FEED_URL="https://repo.cloudradar.io/windows/cagent/feed/plus-rolling"
endif
ifeq ($(RELEASE_MODE),stable)
  SELF_UPDATES_FEED_URL="https://repo.cloudradar.io/windows/cagent/feed/stable"
  PROPRIETARY_SELF_UPDATES_FEED_URL="https://repo.cloudradar.io/windows/cagent/feed/plus-stable"
endif

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARIES=cagent csender jobmon

all: test build

build:
	$(foreach BINARY,$(BINARIES),$(GOBUILD) -o $(BINARY) -v ./cmd/$(BINARY)/...;)

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARIES)

goreleaser-precheck:
	@if [ -z ${SELF_UPDATES_FEED_URL} ]; then echo "SELF_UPDATES_FEED_URL is empty"; exit 1; fi
	@if [ -z ${PROPRIETARY_SELF_UPDATES_FEED_URL} ]; then echo "PROPRIETARY_SELF_UPDATES_FEED_URL is empty"; exit 1; fi

goreleaser-rm-dist: goreleaser-precheck
	GORELEASER_CURRENT_TAG=$(GORELEASER_CURRENT_TAG) SELF_UPDATES_FEED_URL=$(SELF_UPDATES_FEED_URL) PROPRIETARY_SELF_UPDATES_FEED_URL=$(PROPRIETARY_SELF_UPDATES_FEED_URL) goreleaser --rm-dist

goreleaser-snapshot: goreleaser-precheck
	GORELEASER_CURRENT_TAG=$(GORELEASER_CURRENT_TAG) SELF_UPDATES_FEED_URL=$(SELF_UPDATES_FEED_URL) PROPRIETARY_SELF_UPDATES_FEED_URL=$(PROPRIETARY_SELF_UPDATES_FEED_URL) goreleaser --snapshot --rm-dist

goimports:
	goimports -l $$(find . -type f -name '*.go' -not -path "./vendor/*")

docker-goreleaser: goreleaser-precheck
	docker run -it --rm --privileged \
		-v ${PWD}:${PROJECT_DIR} \
		-v $(go env GOCACHE):/root/.cache/go-build \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w ${PROJECT_DIR} \
		-e SELF_UPDATES_FEED_URL=$(SELF_UPDATES_FEED_URL) \
		-e PROPRIETARY_SELF_UPDATES_FEED_URL=$(SELF_UPDATES_FEED_URL) \
		goreleaser/goreleaser:v0.126 --snapshot --rm-dist --skip-publish

docker-golangci-lint:
	docker run -it --rm \
		-v ${PWD}:${PROJECT_DIR} \
		-w ${PROJECT_DIR} \
		golangci/golangci-lint:v1.17 golangci-lint -c .golangci.yml run

 # Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=cagent

all: test build
build: 
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/cagent/...

test: 
	$(GOTEST) -v ./...

clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/cagent/...
	./$(BINARY_NAME)

ci: goreleaser-rm-dist windows-sign

goreleaser-rm-dist:
	goreleaser --rm-dist

goreleaser-snapshot:
	goreleaser --snapshot

goimports:
	goimports -l $$(find . -type f -name '*.go' -not -path "./vendor/*")

windows-sign:
	# Create remote build dir
	ssh -i /tmp/id_win_ssh -p 24481 -oStrictHostKeyChecking=no hero@144.76.9.139 mkdir -p /cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}/dist
	# Copy exe files to Windows VM for bundingling and signing
	scp -i /tmp/id_win_ssh -P 24481 -oStrictHostKeyChecking=no /go/src/github.com/cloudradar-monitoring/cagent/dist/windows_386/cagent.exe hero@144.76.9.139:/cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}/dist/cagent_386.exe
	scp -i /tmp/id_win_ssh -P 24481 -oStrictHostKeyChecking=no /go/src/github.com/cloudradar-monitoring/cagent/dist/windows_amd64/cagent.exe hero@144.76.9.139:/cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}/dist/cagent_64.exe
	# Copy other build dependencies
	scp -i /tmp/id_win_ssh -P 24481 -oStrictHostKeyChecking=no /go/src/github.com/cloudradar-monitoring/cagent/build-win.bat hero@144.76.9.139:/cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}/build-win.bat
	ssh -i /tmp/id_win_ssh -p 24481 -oStrictHostKeyChecking=no hero@144.76.9.139 chmod +x /cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}/build-win.bat
	scp -r -i /tmp/id_win_ssh -P 24481 -oStrictHostKeyChecking=no /go/src/github.com/cloudradar-monitoring/cagent/pkg-scripts/  hero@144.76.9.139:/cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}
	scp -r -i /tmp/id_win_ssh -P 24481 -oStrictHostKeyChecking=no /go/src/github.com/cloudradar-monitoring/cagent/resources/  hero@144.76.9.139:/cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}
	scp -i /tmp/id_win_ssh -P 24481 -oStrictHostKeyChecking=no /go/src/github.com/cloudradar-monitoring/cagent/example.config.toml hero@144.76.9.139:/cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}/example.config.toml
	scp -i /tmp/id_win_ssh -P 24481 -oStrictHostKeyChecking=no /go/src/github.com/cloudradar-monitoring/cagent/README.md hero@144.76.9.139:/cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}/README.md
	scp -i /tmp/id_win_ssh -P 24481 -oStrictHostKeyChecking=no /go/src/github.com/cloudradar-monitoring/cagent/LICENSE hero@144.76.9.139:/cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}/LICENSE
	scp -i /tmp/id_win_ssh -P 24481 -oStrictHostKeyChecking=no /go/src/github.com/cloudradar-monitoring/cagent/wix.json hero@144.76.9.139:/cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}/wix.json
	# Trigger msi creating
	ssh -i /tmp/id_win_ssh -p 24481 -oStrictHostKeyChecking=no hero@144.76.9.139 /cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}/build-win.bat ${CIRCLE_BUILD_NUM} ${CIRCLE_TAG}
	# Trigger signing
	ssh -i /tmp/id_win_ssh -p 24481 -oStrictHostKeyChecking=no hero@144.76.9.139 curl http://localhost:8088/?file=cagent_32.msi
	echo $$?
	sleep 10
	ssh -i /tmp/id_win_ssh -p 24481 -oStrictHostKeyChecking=no hero@144.76.9.139 curl http://localhost:8080/?file=cagent_64.msi
	# Copy msi files back to build machine
	scp -i /tmp/id_win_ssh -P 24481 -oStrictHostKeyChecking=no hero@144.76.9.139:/cygdrive/C/Users/hero/ci/cagent_32.msi /go/src/github.com/cloudradar-monitoring/cagent/dist/cagent_386.msi
	scp -i /tmp/id_win_ssh -P 24481 -oStrictHostKeyChecking=no hero@144.76.9.139:/cygdrive/C/Users/hero/ci/cagent_64.msi /go/src/github.com/cloudradar-monitoring/cagent/dist/cagent_64.msi
	# Add files to Github release
	github-release upload --user cloudradar-monitoring --repo cagent --tag ${CIRCLE_TAG} --name "cagent_${CIRCLE_TAG}_Windows_386.msi" --file "/go/src/github.com/cloudradar-monitoring/cagent/dist/cagent_386.msi"
	github-release upload --user cloudradar-monitoring --repo cagent --tag ${CIRCLE_TAG} --name "cagent_${CIRCLE_TAG}_Windows_x86_64.msi" --file "/go/src/github.com/cloudradar-monitoring/cagent/dist/cagent_64.msi"

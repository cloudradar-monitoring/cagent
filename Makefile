.PHONY: synology-spk

PROJECT_DIR=/go/src/github.com/cloudradar-monitoring/cagent
WIN_BUILD_MACHINE_CI_DIR=/cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}
WIN_BUILD_MACHINE_CI_DIR_PROPRIETARY=/cygdrive/C/Users/hero/ci/cagent_ci/build_msi/${CIRCLE_BUILD_NUM}_proprietary
WIN_BUILD_MACHINE_PROPRIETARY_DEPS_DIR=/cygdrive/C/Users/hero/ci/cagent_ci/proprietary_deps
WIN_BUILD_MACHINE_AUTH=hero@144.76.9.139
SCP_WIN_BUILD_MACHINE_OPTIONS=-P 24481 -oStrictHostKeyChecking=no
SSH_WIN_BUILD_MACHINE_OPTIONS=-p 24481 -oStrictHostKeyChecking=no

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

ci: goreleaser-rm-dist windows-sign

aptly:
	# Create remote work dir
	ssh -p 24480 -oStrictHostKeyChecking=no cr@repo.cloudradar.io mkdir -p /home/cr/work/aptly/cagent_${CIRCLE_BUILD_NUM}
	# Upload deb files
	rsync -e 'ssh -oStrictHostKeyChecking=no -p 24480' --recursive ${PROJECT_DIR}/dist/*.deb --exclude "*_armv6.deb"  cr@repo.cloudradar.io:/home/cr/work/aptly/cagent_${CIRCLE_BUILD_NUM}/
	# Trigger repository update
	ssh -p 24480 -oStrictHostKeyChecking=no cr@repo.cloudradar.io /home/cr/work/aptly/update_repo.sh /home/cr/work/aptly/cagent_${CIRCLE_BUILD_NUM} ${CIRCLE_TAG}

createrepo:
	# Create remote work dir
	ssh -p 24480 -oStrictHostKeyChecking=no cr@repo.cloudradar.io mkdir -p /home/cr/work/rpm/cagent_${CIRCLE_BUILD_NUM}
	# Upload rpm files
	rsync -e 'ssh -oStrictHostKeyChecking=no -p 24480' --recursive ${PROJECT_DIR}/dist/*.rpm  cr@repo.cloudradar.io:/home/cr/work/rpm/cagent_${CIRCLE_BUILD_NUM}/
	# Trigger repository update
	ssh -p 24480 -oStrictHostKeyChecking=no cr@repo.cloudradar.io /home/cr/work/rpm/update_repo_cagent.sh /home/cr/work/rpm/cagent_${CIRCLE_BUILD_NUM} ${CIRCLE_TAG}

goreleaser-rm-dist:
	goreleaser --rm-dist

goreleaser-snapshot:
	goreleaser --snapshot

goimports:
	goimports -l $$(find . -type f -name '*.go' -not -path "./vendor/*")

windows-sign:
	# BUILD MIT-licensed MSI:
	# Create remote build dir
	ssh ${SSH_WIN_BUILD_MACHINE_OPTIONS} ${WIN_BUILD_MACHINE_AUTH} mkdir -p ${WIN_BUILD_MACHINE_CI_DIR}/dist
	# Copy exe files to Windows VM for bundling and signing
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/dist/cagent_windows_amd64/cagent.exe ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR}/dist/cagent_64.exe
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/dist/csender_windows_amd64/csender.exe ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR}/dist/csender_64.exe
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/dist/jobmon_windows_amd64/jobmon.exe ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR}/dist/jobmon_64.exe
	# Copy other build dependencies
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/build-win.bat ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR}/build-win.bat
	ssh ${SSH_WIN_BUILD_MACHINE_OPTIONS} ${WIN_BUILD_MACHINE_AUTH} chmod +x ${WIN_BUILD_MACHINE_CI_DIR}/build-win.bat
	scp -r ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/pkg-scripts/ ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR}
	scp -r ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/resources/ ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR}
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/example.config.toml ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR}/example.config.toml
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/README.md ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR}/README.md
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/LICENSE ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR}/LICENSE
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/wix.json ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR}/wix.json
	# Trigger MSI build
	ssh ${SSH_WIN_BUILD_MACHINE_OPTIONS} ${WIN_BUILD_MACHINE_AUTH} ${WIN_BUILD_MACHINE_CI_DIR}/build-win.bat ${CIRCLE_BUILD_NUM} ${CIRCLE_TAG}
	# Trigger signing
	ssh ${SSH_WIN_BUILD_MACHINE_OPTIONS} ${WIN_BUILD_MACHINE_AUTH} curl -s -S -f http://localhost:8080/?file=cagent_64.msi
	# Copy MSI back to build machine
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${WIN_BUILD_MACHINE_AUTH}:/cygdrive/C/Users/hero/ci/cagent_64.msi ${PROJECT_DIR}/dist/cagent_64.msi
	# Add files to Github release
	github-release upload --user cloudradar-monitoring --repo cagent --tag ${CIRCLE_TAG} --name "cagent_${CIRCLE_TAG}_Windows_x86_64.msi" --file "${PROJECT_DIR}/dist/cagent_64.msi"

	# BUILD PROPRIETARY MSI:
	# Create remote build dir
	ssh ${SSH_WIN_BUILD_MACHINE_OPTIONS} ${WIN_BUILD_MACHINE_AUTH} mkdir -p ${WIN_BUILD_MACHINE_CI_DIR_PROPRIETARY}/dist
	# Copy exe files to Windows VM for bundling and signing
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/dist/cagent_proprietary_windows_amd64/cagent.exe ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR_PROPRIETARY}/dist/cagent_64.exe
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/dist/csender_proprietary_windows_amd64/csender.exe ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR_PROPRIETARY}/dist/csender_64.exe
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/dist/jobmon_proprietary_windows_amd64/jobmon.exe ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR_PROPRIETARY}/dist/jobmon_64.exe
	# Copy other build dependencies
	scp -r ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/pkg-scripts/ ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR_PROPRIETARY}
	scp -r ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/resources/ ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR_PROPRIETARY}
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/example.config.toml ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR_PROPRIETARY}/example.config.toml
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${PROJECT_DIR}/README.md ${WIN_BUILD_MACHINE_AUTH}:${WIN_BUILD_MACHINE_CI_DIR_PROPRIETARY}/README.md
	# Trigger MSI build
	ssh ${SSH_WIN_BUILD_MACHINE_OPTIONS} ${WIN_BUILD_MACHINE_AUTH} cp -r ${WIN_BUILD_MACHINE_PROPRIETARY_DEPS_DIR}/. ${WIN_BUILD_MACHINE_CI_DIR_PROPRIETARY}
	ssh ${SSH_WIN_BUILD_MACHINE_OPTIONS} ${WIN_BUILD_MACHINE_AUTH} ${WIN_BUILD_MACHINE_CI_DIR_PROPRIETARY}/build-proprietary-win.bat ${CIRCLE_BUILD_NUM}_proprietary ${CIRCLE_TAG}
	# Trigger signing
	ssh ${SSH_WIN_BUILD_MACHINE_OPTIONS} ${WIN_BUILD_MACHINE_AUTH} curl -s -S -f http://localhost:8080/?file=cagent_64.msi
	# Copy MSI back to build machine
	scp ${SCP_WIN_BUILD_MACHINE_OPTIONS} ${WIN_BUILD_MACHINE_AUTH}:/cygdrive/C/Users/hero/ci/cagent_64.msi ${PROJECT_DIR}/dist/cagent_proprietary_64.msi
	# Upload proprietary MSI files to cloudradar package server
	chmod a+r ${PROJECT_DIR}/dist/cagent_proprietary_64.msi
	scp -P 24480 -oStrictHostKeyChecking=no -p ${PROJECT_DIR}/dist/cagent_proprietary_64.msi cr@repo.cloudradar.io:/var/repos/cloudradar/windows/cagent/cagent-plus_${CIRCLE_TAG}_Windows_64.msi
	# Update page with the link to latest stable MSI
	ssh -p 24480 -oStrictHostKeyChecking=no cr@repo.cloudradar.io /home/cr/update-latest.sh

synology-spk:
	## should call goreleaser before, as it depends on the dist generated by it
	cd synology-spk && ./create_spk.sh ${CIRCLE_TAG}
	# Add files to Github release
	github-release upload --user cloudradar-monitoring --repo cagent --tag ${CIRCLE_TAG} --name "cagent_${CIRCLE_TAG}_synology_amd64.spk" --file "${PROJECT_DIR}/dist/cagent_${CIRCLE_TAG}_synology_amd64.spk"
	github-release upload --user cloudradar-monitoring --repo cagent --tag ${CIRCLE_TAG} --name "cagent_${CIRCLE_TAG}_synology_armv7.spk" --file "${PROJECT_DIR}/dist/cagent_${CIRCLE_TAG}_synology_armv7.spk"
	github-release upload --user cloudradar-monitoring --repo cagent --tag ${CIRCLE_TAG} --name "cagent_${CIRCLE_TAG}_synology_arm64.spk" --file "${PROJECT_DIR}/dist/cagent_${CIRCLE_TAG}_synology_arm64.spk"

docker-goreleaser:
	docker run -it --rm --privileged \
		-v ${PWD}:${PROJECT_DIR} \
		-v $(go env GOCACHE):/root/.cache/go-build \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w ${PROJECT_DIR} \
		goreleaser/goreleaser:v0.111 --snapshot --rm-dist --skip-publish

docker-golangci-lint:
	docker run -it --rm \
		-v ${PWD}:${PROJECT_DIR} \
		-w ${PROJECT_DIR} \
		golangci/golangci-lint:v1.17 golangci-lint -c .golangci.yml run

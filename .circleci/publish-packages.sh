#!/usr/bin/env bash

set -xe

ssh_ci() {
  ssh -p 24480 -oStrictHostKeyChecking=no ci@repo.cloudradar.io "$@"
}

ssh_cr() {
  ssh -p 24480 -oStrictHostKeyChecking=no cr@repo.cloudradar.io "$@"
}

github_upload() {
  github-release upload --user cloudradar-monitoring --repo cagent --tag ${CIRCLE_TAG} "$@"
}

if [ -z ${RELEASE_MODE} ]; then
  echo "RELEASE_MODE env variable is empty"
  exit 1
fi

PROJECT_NAME=github.com/cloudradar-monitoring/cagent
PROJECT_DIR=/go/src/${PROJECT_NAME}
WORK_DIR=/home/ci/buffer/${CIRCLE_BUILD_NUM}

# create build dir structure
ssh_ci mkdir -p ${WORK_DIR}/deb
ssh_ci mkdir -p ${WORK_DIR}/rpm
ssh_ci mkdir -p ${WORK_DIR}/msi

# send packages
rsync -e 'ssh -oStrictHostKeyChecking=no -p 24480' ${PROJECT_DIR}/dist/*.deb --exclude "*_armv6.deb"  ci@repo.cloudradar.io:${WORK_DIR}/deb/
rsync -e 'ssh -oStrictHostKeyChecking=no -p 24480' ${PROJECT_DIR}/dist/*.rpm  ci@repo.cloudradar.io:${WORK_DIR}/rpm/
rsync -e 'ssh -oStrictHostKeyChecking=no -p 24480' -a --files-from=${PROJECT_DIR}/.circleci/rsync-msi-build.list ${PROJECT_DIR} ci@repo.cloudradar.io:${WORK_DIR}/msi/

# publish to DEB and RPM repo
if [ ${RELEASE_MODE} = "stable" ]; then
  ssh_cr /home/cr/work/aptly/update_repo.sh ${WORK_DIR}/deb ${CIRCLE_TAG}
  ssh_cr /home/cr/work/rpm/update_repo_cagent.sh ${WORK_DIR}/rpm ${CIRCLE_TAG}
fi

# trigger MSI build and sign
ssh_cr /home/cr/work/msi/cagent_build_and_sign_msi.sh ${WORK_DIR}/msi ${CIRCLE_BUILD_NUM} ${CIRCLE_TAG}

# copy signed MSI back
scp -P 24480 -oStrictHostKeyChecking=no ci@repo.cloudradar.io:${WORK_DIR}/msi/cagent_64.msi ${PROJECT_DIR}/dist/cagent_64.msi

# publish built files to Github
github_upload --name "cagent_${CIRCLE_TAG}_Windows_x86_64.msi" --file "${PROJECT_DIR}/dist/cagent_64.msi"
github_upload --name "cagent_${CIRCLE_TAG}_synology_amd64.spk" --file "${PROJECT_DIR}/dist/cagent_${CIRCLE_TAG}_synology_amd64.spk"
github_upload --name "cagent_${CIRCLE_TAG}_synology_armv7.spk" --file "${PROJECT_DIR}/dist/cagent_${CIRCLE_TAG}_synology_armv7.spk"
github_upload --name "cagent_${CIRCLE_TAG}_synology_arm64.spk" --file "${PROJECT_DIR}/dist/cagent_${CIRCLE_TAG}_synology_arm64.spk"

# fetch release changelog so we can preserve it when releasing
CHANGELOGRAW=$(curl -H "Authorization: token ${GITHUB_TOKEN}" https://api.github.com/repos/cloudradar-monitoring/cagent/releases | jq ".[0].body")

# mark release as published
PRERELEASE="--pre-release"
if [ ${RELEASE_MODE} = "stable" ]; then
  PRERELEASE=
fi
echo -e ${CHANGELOGRAW} | sed -e 's/^"//' -e 's/"$//' | github-release edit --user cloudradar-monitoring --repo cagent --tag ${CIRCLE_TAG} ${PRERELEASE} --description -

# update MSI repo
ssh_cr /home/cr/work/msi/cagent_publish.sh ${WORK_DIR}/msi ${CIRCLE_TAG} ${RELEASE_MODE}

# remove work dir
ssh_ci rm -rf ${WORK_DIR}
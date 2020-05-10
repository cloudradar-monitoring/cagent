#!/usr/bin/env bash

set -xe

ssh_cr() {
  ssh -p 24480 -oStrictHostKeyChecking=no cr@repo.cloudradar.io "$@"
}

ssh_cr /home/cr/work/msi/feed_delete.sh cagent rolling ${CIRLE_TAG}
ssh_cr /home/cr/work/msi/feed_delete.sh cagent plus-rolling ${CIRLE_TAG}

ssh_cr /home/cr/work/msi/feed_delete.sh cagent stable ${CIRLE_TAG}
ssh_cr /home/cr/work/msi/feed_delete.sh cagent plus-stable ${CIRLE_TAG}

github-release delete --user cloudradar-monitoring --repo cagent --tag ${CIRCLE_TAG}

ssh_cr /home/cr/work/msi/cagent_update_latest.sh

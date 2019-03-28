#!/bin/sh

# rpm executes this script with the number of previous known versions. It allows us to determine if we upgrading or installing first time:
# Removing during upgrade:         1 or higher
# Remove last version of package:  0
prevVersionsCount=$1

# we need to uninstall service only if we are removing the last packages
if [[ ${prevVersionsCount} -lt 1 ]] ; then
  /usr/bin/cagent -u || true
fi
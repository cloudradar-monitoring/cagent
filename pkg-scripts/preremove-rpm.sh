#!/bin/sh

# Removing during upgrade:         1 or higher
# Remove last version of package:  0
versionsCount=$1

# we need to uninstall service only if we are removing the last packages
if [ ${versionsCount} -lt 1 ]; then
  /usr/bin/cagent -u || true
fi
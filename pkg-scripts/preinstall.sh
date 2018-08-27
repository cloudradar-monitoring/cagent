#!/bin/sh

# check that owner group exists
if [ -z `getent group cagent` ]; then
  groupadd cagent
fi

# check that user exists
if [ -z `getent passwd cagent` ]; then
  useradd  --gid cagent --system --shell /bin/false cagent
fi
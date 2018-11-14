#!/bin/sh
/usr/bin/cagent -s cagent -c /etc/cagent/cagent.conf
/usr/bin/cagent -t || true

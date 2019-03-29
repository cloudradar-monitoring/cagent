#!/bin/sh
/usr/bin/cagent -y -s cagent -c /etc/cagent/cagent.conf
/usr/bin/cagent -t || true

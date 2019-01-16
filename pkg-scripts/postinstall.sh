#!/bin/sh

# WARN: path to config is referenced by plesk extension
#       when enabling/disabling plugin thus plesk extension
#       has to be updated prior to cagent
/usr/bin/cagent -s cagent -c /etc/cagent/cagent.conf
/usr/bin/cagent -t || true

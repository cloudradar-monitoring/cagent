#!/usr/bin/env bash
source /pkgscripts/include/pkg_util.sh

package="cagent-dev"
version="0.0.1000"
displayname="cagent-dev"
maintainer="Cloudradar GmbH"
arch="$(pkg_get_unified_platform)"
description="this is a dev package"
[ "$(caller)" != "0 NULL" ] && return 0
pkg_dump_info

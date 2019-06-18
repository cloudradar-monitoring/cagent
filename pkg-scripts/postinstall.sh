#!/bin/sh

CONFIG_PATH=/etc/cagent/cagent.conf

if [ "$1" = configure ]; then
    # $2 contains previous version number
    if [ -z "$2" ]; then # fresh install
        /usr/bin/cagent -y -s cagent -c ${CONFIG_PATH}
    else # package update
        serviceStatus=`/usr/bin/cagent -y -service_status -c ${CONFIG_PATH}`
        echo "current service status: $serviceStatus."

        if [ "$serviceStatus" != stopped ]; then
            echo "stopping service..."
            /usr/bin/cagent -service_stop || true
        fi

        echo "upgrading service unit... "
        /usr/bin/cagent -y -s cagent -service_upgrade -c ${CONFIG_PATH}

        # restart only if it was active before
        if [ "$serviceStatus" != stopped ]; then
            echo "restarting service... "
            /usr/bin/cagent -y -service_restart -c ${CONFIG_PATH}
        fi
    fi
fi

/usr/bin/cagent -t || true

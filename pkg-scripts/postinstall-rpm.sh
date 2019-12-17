#!/bin/sh

CONFIG_PATH=/etc/cagent/cagent.conf

# Install the first time:	1
# Upgrade:	2 or higher (depending on the number of versions installed)
versionsCount=$1

if [ ${versionsCount} == 1 ]; then # fresh install
    chown cagent:cagent -R /var/lib/cagent
    chmod 6777 /var/lib/cagent/jobmon
    /usr/bin/cagent -y -s cagent -c ${CONFIG_PATH}
else # package update
    serviceStatus=`/usr/bin/cagent -y -service_status -c ${CONFIG_PATH}`
    echo "current service status: $serviceStatus."

    case "$serviceStatus" in
        unknown|failed)
            echo "trying to repair service..."
            /usr/bin/cagent -u || true
            /usr/bin/cagent -y -s cagent -c ${CONFIG_PATH}
            ;;

        running|stopped)
            # try to upgrade service unit config

            if [ "$serviceStatus" = running ]; then
                echo "stopping service..."
                /usr/bin/cagent -service_stop || true
            fi

            echo "upgrading service unit... "
            /usr/bin/cagent -y -s cagent -service_upgrade -c ${CONFIG_PATH}

            # restart only if it was active before
            if [ "$serviceStatus" = running ]; then
                echo "starting service... "
                /usr/bin/cagent -y -service_start -c ${CONFIG_PATH}
            fi
            ;;

        *)
            echo "unknown service status. Exiting..."
            exit 1
            ;;
    esac
fi

/usr/bin/cagent -t || true

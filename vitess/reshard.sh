#!/bin/bash

# Enable tty for Windows users using git-bash or cygwin
if [[ "$OSTYPE" == "msys" ]]; then
        # Lightweight shell and GNU utilities compiled for Windows (part of MinGW)
        tty=winpty
fi

exec $tty docker-compose exec ${CS:-vtctld} /vt/bin/vtctlclient -server vtctld:15999 Reshard --source_shards '0' --target_shards '-80,80-' Create conversation_keyspace.reshard

sleep 10

exec $tty docker-compose exec ${CS:-vtctld} /vt/bin/vtctlclient -server vtctld:15999 Reshard -- --tablet_types=rdonly,replica SwitchTraffic conversation_keyspace.reshard
exec $tty docker-compose exec ${CS:-vtctld} /vt/bin/vtctlclient -server vtctld:15999 Reshard -- --tablet_types=primary SwitchTraffic conversation_keyspace.reshard


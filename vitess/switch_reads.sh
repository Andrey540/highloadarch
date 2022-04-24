#!/bin/bash

# Enable tty for Windows users using git-bash or cygwin
if [[ "$OSTYPE" == "msys" ]]; then
        # Lightweight shell and GNU utilities compiled for Windows (part of MinGW)
        tty=winpty
fi

exec $tty docker-compose exec ${CS:-vtctld} /vt/bin/vtctlclient -server vtctld:15999 Reshard -- --tablet_types=rdonly,replica SwitchTraffic conversation_keyspace.reshard


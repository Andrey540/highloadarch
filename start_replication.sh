#!/usr/bin/env bash

docker exec mysql-node-1 mysql -uroot \
  -e "SET @@GLOBAL.group_replication_bootstrap_group=1;" \
  -e "create user IF NOT EXISTS repl@'%';" \
  -e "GRANT REPLICATION SLAVE ON *.* TO repl@'%';" \
  -e "flush privileges;" \
  -e "change master to master_user='root' for channel 'group_replication_recovery';" \
  -e "START GROUP_REPLICATION;" \
  -e "SET @@GLOBAL.group_replication_bootstrap_group=0;"

docker exec mysql-node-2 mysql -uroot \
  -e "change master to master_user='repl' for channel 'group_replication_recovery';" \
  -e "START GROUP_REPLICATION;"

docker exec mysql-node-3 mysql -uroot \
  -e "change master to master_user='repl' for channel 'group_replication_recovery';" \
  -e "START GROUP_REPLICATION;"
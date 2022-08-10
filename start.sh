#!/usr/bin/env bash

docker-compose up -d

echo -e 'Waiting for containers...'

sleep 10

echo -e 'Starting redis replication...'

CHECK=$(docker inspect -f '{{.State.Status}}' redis-node-2)
while [ "$CHECK" != "running" ]
do
    sleep 2
    echo -e 'Waiting for redis-node-2 getting ready...'
done

CHECK=$(docker inspect -f '{{.State.Status}}' redis-node-3)
while [ "$CHECK" != "running" ]
do
    sleep 2
    echo -e 'Waiting for redis-node-3 getting ready...'
done

docker exec -it redis-node-2 redis-cli SLAVEOF redis-node-1 6379
docker exec -it redis-node-3 redis-cli SLAVEOF redis-node-1 6379



echo -e 'Starting MySQL replication...'

CHECK=$(docker inspect -f '{{.State.Status}}' mysql-node-1)
while [ "$CHECK" != "running" ]
do
    sleep 2
    echo -e 'Waiting for mysql-node-1 getting ready...'
done

CHECK=$(docker inspect -f '{{.State.Status}}' mysql-node-2)
while [ "$CHECK" != "running" ]
do
    sleep 2
    echo -e 'Waiting for mysql-node-2 getting ready...'
done

CHECK=$(docker inspect -f '{{.State.Status}}' mysql-node-3)
while [ "$CHECK" != "running" ]
do
    sleep 2
    echo -e 'Waiting for mysql-node-3 getting ready...'
done

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



echo -e 'Starting Tarantool replication...'

CHECK=$(docker inspect -f '{{.State.Status}}' mysql-post)
while [ "$CHECK" != "running" ]
do
    sleep 2
    echo -e 'Waiting for mysql-post getting ready...'
done

CHECK=$(docker inspect -f '{{.State.Status}}' tarantool)
while [ "$CHECK" != "running" ]
do
    sleep 2
    echo -e 'Waiting for tarantool getting ready...'
done

CHECK=$(docker inspect -f '{{.State.Status}}' tarantool-replicator)
while [ "$CHECK" != "running" ]
do
    sleep 2
    echo -e 'Waiting for tarantool-replicator getting ready...'
done

docker exec tarantool-replicator /bin/bash -c "systemctl start replicatord"
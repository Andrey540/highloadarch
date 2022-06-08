# Отчёт по домашнему заданию №9

## Описание

Для реализации отказоустойчивости работы с репликой MySQL использован haproxy. Его конфиги находятся по пути haproxy/haproxy.cnf

Для реализации отказоустойчивости приложения использован nginx. Отказоустойчивым сделан микросервис socialnetwork,
тк он является основной точкой отказа. Конфиги nginx находятся по пути nginx/nginx.conf

Поднимаем контейнеры.
```bash
docker-compose up
```

И запустить репликацию.
```bash
./start_replication.sh
```

Запускаем нагрузку
```bash
wrk -t1 -c1 -d60s --latency --script="config.lua" http://127.0.0.1:8881/app/profile
```

Перед запуском нагрузки нужно зайти в приложение и вытащить куку.
Файл config.lau имеет следующее содержание
```bash
wrk.headers["Cookie"] = "otussid=wgErKziZiUCmviiWtp3aDF"
```

Далее останавливаем реплику
```bash
sudo pkill -9 -f mysql-node-2
```

Убеждаемся что система осталась работоспособной. Можно посомтреть только ответы nginx, статусы 200.
НИже приведены логи в момент падение второй ноды mysql-node-2. Можно видеть, что микросервис user потерял соединение, а затем заново пореподключился.
mysql-node-2 завершила свою работу, а ноды mysql-node-1 mysql-node-3 потеряли mysql-node-2. Также haproxy обнаружил потерю.
Однако можно видеть, что nginx отдаёт статус 200 и микросервис user работает штатно.
```bash
nginx                   | 172.18.0.1 - - [08/Jun/2022:17:33:47 +0000] "GET /app/profile HTTP/1.1" 200 3569 "-" "-" "-"
nginx                   | 172.18.0.1 - - [08/Jun/2022:17:33:47 +0000] "GET /app/profile HTTP/1.1" 200 3569 "-" "-" "-"
socialnetwork-1         | http: 2022/06/08 17:33:47 RequestID: 26c932d0-e751-11ec-9fea-0242ac120012 GET /app/profile 172.18.0.20:51256 
socialnetwork-1         | http: 2022/06/08 17:33:47 Response: 
socialnetwork-1         | 
user                    | http: 2022/06/08 17:33:47 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:1.655218ms method:GetProfile requestID:26c932d0-e751-11ec-9fea-0242ac120012] call finished
user                    | http: 2022/06/08 17:33:47 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:1.509448ms method:ListFriends requestID:26c932d0-e751-11ec-9fea-0242ac120012] call finished
user                    | [mysql] 2022/06/08 17:33:47 packets.go:123: closing bad idle connection: EOF
user                    | [mysql] 2022/06/08 17:33:47 connection.go:173: driver: bad connection
haproxy                 | Connect from 172.18.0.11:49194 to 172.18.0.12:3306 (mysql-cluster/TCP)
user                    | http: 2022/06/08 17:33:47 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:4.830767ms method:GetProfile requestID:26f7ee89-e751-11ec-a9f2-0242ac120013] call finished
user                    | http: 2022/06/08 17:33:47 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:1.272594ms method:ListFriends requestID:26f7ee89-e751-11ec-a9f2-0242ac120013] call finished
socialnetwork-2         | http: 2022/06/08 17:33:47 RequestID: 26f7ee89-e751-11ec-a9f2-0242ac120013 GET /app/profile 172.18.0.20:50120 
socialnetwork-2         | http: 2022/06/08 17:33:47 Response: 
socialnetwork-2         | 
nginx                   | 172.18.0.1 - - [08/Jun/2022:17:33:47 +0000] "GET /app/profile HTTP/1.1" 200 3569 "-" "-" "-"
mysql-node-1            | 2022-06-08T17:33:47.737956Z 0 [ERROR] [MY-011735] [Repl] Plugin group_replication reported: '[GCS] Error retrieving server information.'
mysql-node-2 exited with code 137
mysql-node-3            | 2022-06-08T17:33:47.803186Z 0 [ERROR] [MY-011735] [Repl] Plugin group_replication reported: '[GCS] Error retrieving server information.'
mysql-node-1            | 2022-06-08T17:33:47.838184Z 0 [ERROR] [MY-011735] [Repl] Plugin group_replication reported: '[GCS] Error retrieving server information.'
mysql-node-3            | 2022-06-08T17:33:47.904994Z 0 [ERROR] [MY-011735] [Repl] Plugin group_replication reported: '[GCS] Error retrieving server information.'
user                    | http: 2022/06/08 17:33:47 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:1.738672ms method:GetProfile requestID:2727134f-e751-11ec-9fea-0242ac120012] call finished
user                    | http: 2022/06/08 17:33:47 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:1.51795ms method:ListFriends requestID:2727134f-e751-11ec-9fea-0242ac120012] call finished
socialnetwork-1         | http: 2022/06/08 17:33:47 RequestID: 2727134f-e751-11ec-9fea-0242ac120012 GET /app/profile 172.18.0.20:51258 
socialnetwork-1         | http: 2022/06/08 17:33:47 Response: 
socialnetwork-1         | 
nginx                   | 172.18.0.1 - - [08/Jun/2022:17:33:47 +0000] "GET /app/profile HTTP/1.1" 200 3569 "-" "-" "-"
user                    | http: 2022/06/08 17:33:48 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:1.681705ms method:GetProfile requestID:275612c4-e751-11ec-a9f2-0242ac120013] call finished
user                    | http: 2022/06/08 17:33:48 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:1.444776ms method:ListFriends requestID:275612c4-e751-11ec-a9f2-0242ac120013] call finished
socialnetwork-2         | http: 2022/06/08 17:33:48 RequestID: 275612c4-e751-11ec-a9f2-0242ac120013 GET /app/profile 172.18.0.20:50122 
socialnetwork-2         | http: 2022/06/08 17:33:48 Response: 
socialnetwork-2         | 
haproxy                 | Server mysql-cluster/node2 is DOWN, reason: Layer4 timeout, check duration: 2001ms. 1 active and 0 backup servers left. 0 sessions active, 0 requeued, 0 remaining in queue.
haproxy                 | [WARNING]  (8) : Server mysql-cluster/node2 is DOWN, reason: Layer4 timeout, check duration: 2001ms. 1 active and 0 backup servers left. 0 sessions active, 0 requeued, 0 remaining in queue.
```

Аналогичным образом проделываем для приложения. В данном случае имеется два контейнера - socialnetwork-1 и socialnetwork-2.

Запускаем нагрузку
```bash
wrk -t1 -c1 -d60s --latency --script="config.lua" http://127.0.0.1:8881/app/profile
```

Командой ниже находим PID процессов
```bash
ps aux | grep socialnetwork
```

И останавливаем один из них
```bash
sudo kill -9 PID
```

Ниже приведены логи работы приложения.
Из логов видно, что контейнер socialnetwork-1 остановил свою работу аварийно. Перед этим nginx обнаружил пропажу контейнера.
Тем не менее приложение продолжает работать корректно, nginx отдаёт 200 статус. Микросервис user так же отрабатывает.
```bash
user                    | http: 2022/06/08 17:56:29 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:515.001µs method:GetProfile requestID:52c328eb-e754-11ec-85d2-0242ac120013] call finished
user                    | http: 2022/06/08 17:56:29 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:318.207µs method:ListFriends requestID:52c328eb-e754-11ec-85d2-0242ac120013] call finished
socialnetwork-2         | http: 2022/06/08 17:56:29 RequestID: 52c328eb-e754-11ec-85d2-0242ac120013 GET /app/profile 172.18.0.20:50176 
socialnetwork-2         | http: 2022/06/08 17:56:29 Response: 
socialnetwork-2         | 
nginx                   | 172.18.0.1 - - [08/Jun/2022:17:56:29 +0000] "GET /app/profile HTTP/1.1" 200 3569 "-" "-" "-"
user                    | http: 2022/06/08 17:56:29 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:1.495655ms method:GetProfile requestID:52f14c6d-e754-11ec-a79e-0242ac120012] call finished
user                    | http: 2022/06/08 17:56:29 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:1.587136ms method:ListFriends requestID:52f14c6d-e754-11ec-a79e-0242ac120012] call finished
socialnetwork-1         | http: 2022/06/08 17:56:29 RequestID: 52f14c6d-e754-11ec-a79e-0242ac120012 GET /app/profile 172.18.0.20:51310 
socialnetwork-1         | http: 2022/06/08 17:56:29 Response: 
socialnetwork-1         | 
nginx                   | 172.18.0.1 - - [08/Jun/2022:17:56:29 +0000] "GET /app/profile HTTP/1.1" 200 3569 "-" "-" "-"
conversation            | ERROR: logging before flag.Parse: W0608 17:56:30.255303       1 component.go:41] [core] grpc: addrConn.createTransport failed to connect to {vtctld:15999 vtctld:15999 <nil> 0 <nil>}. Err: connection error: desc = "transport: Error while dialing dial tcp: lookup vtctld on 127.0.0.11:53: server misbehaving"
user                    | http: 2022/06/08 17:56:30 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:1.64787ms method:GetProfile requestID:53202c95-e754-11ec-85d2-0242ac120013] call finished
user                    | http: 2022/06/08 17:56:30 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:1.557575ms method:ListFriends requestID:53202c95-e754-11ec-85d2-0242ac120013] call finished
socialnetwork-2         | http: 2022/06/08 17:56:30 RequestID: 53202c95-e754-11ec-85d2-0242ac120013 GET /app/profile 172.18.0.20:50178 
socialnetwork-2         | http: 2022/06/08 17:56:30 Response: 
socialnetwork-2         | 
nginx                   | 172.18.0.1 - - [08/Jun/2022:17:56:30 +0000] "GET /app/profile HTTP/1.1" 200 3569 "-" "-" "-"
nginx                   | 2022/06/08 17:56:30 [error] 24#24: *2 upstream prematurely closed connection while reading response header from upstream, client: 172.18.0.1, server: , request: "GET /app/profile HTTP/1.1", upstream: "http://172.18.0.18:80/app/profile", host: "127.0.0.1:8881"
nginx                   | 2022/06/08 17:56:30 [warn] 24#24: *2 upstream server temporarily disabled while reading response header from upstream, client: 172.18.0.1, server: , request: "GET /app/profile HTTP/1.1", upstream: "http://172.18.0.18:80/app/profile", host: "127.0.0.1:8881"
socialnetwork-1 exited with code 137
user                    | http: 2022/06/08 17:56:30 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:820.764µs method:GetProfile requestID:535aa3ba-e754-11ec-85d2-0242ac120013] call finished
user                    | http: 2022/06/08 17:56:30 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:667.223µs method:ListFriends requestID:535aa3ba-e754-11ec-85d2-0242ac120013] call finished
socialnetwork-2         | http: 2022/06/08 17:56:30 RequestID: 535aa3ba-e754-11ec-85d2-0242ac120013 GET /app/profile 172.18.0.20:50180 
socialnetwork-2         | http: 2022/06/08 17:56:30 Response: 
socialnetwork-2         | 
nginx                   | 172.18.0.1 - - [08/Jun/2022:17:56:30 +0000] "GET /app/profile HTTP/1.1" 200 3569 "-" "-" "-"
user                    | http: 2022/06/08 17:56:30 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:549.015µs method:GetProfile requestID:53890bd9-e754-11ec-85d2-0242ac120013] call finished
user                    | http: 2022/06/08 17:56:30 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:438.684µs method:ListFriends requestID:53890bd9-e754-11ec-85d2-0242ac120013] call finished
socialnetwork-2         | http: 2022/06/08 17:56:30 RequestID: 53890bd9-e754-11ec-85d2-0242ac120013 GET /app/profile 172.18.0.20:50182 
socialnetwork-2         | http: 2022/06/08 17:56:30 Response: 
socialnetwork-2         | 
nginx                   | 172.18.0.1 - - [08/Jun/2022:17:56:30 +0000] "GET /app/profile HTTP/1.1" 200 3569 "-" "-" "-"
user                    | http: 2022/06/08 17:56:31 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:1.000444ms method:GetProfile requestID:53b73f91-e754-11ec-85d2-0242ac120013] call finished
user                    | http: 2022/06/08 17:56:31 map[args:userID:"fe2796fa-e4e6-11ec-9493-0242ac12000c" duration:1.661594ms method:ListFriends requestID:53b73f91-e754-11ec-85d2-0242ac120013] call finished
socialnetwork-2         | http: 2022/06/08 17:56:31 RequestID: 53b73f91-e754-11ec-85d2-0242ac120013 GET /app/profile 172.18.0.20:50184 
socialnetwork-2         | http: 2022/06/08 17:56:31 Response: 
socialnetwork-2         | 
nginx                   | 172.18.0.1 - - [08/Jun/2022:17:56:31 +0000] "GET /app/profile HTTP/1.1" 200 3569 "-" "-" "-"
```
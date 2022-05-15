# Отчёт по домашнему заданию №6

## Описание

Для репликатора был создан отдельный докер образ andrey540/mysql-tarantool-replication.
После ДЗ №5 была изменена схема БД. В таблице news_line мискросервиса post были заменены uuid-ы с типа binary(16) на varchar(36).
Сделано это потому, что репликатор тарантула не поддерживает тип binary(16). Изначально хотел репликацию в тарантул развернуть в user,
но тк binary(16) не поддерживается, то это привело бы к большим переделкам, а также к потере производительности, тк выборка по полю с типом binary(16) наботает быстрее.

Поэтому прежде всего нужно удалить старую базу. Для этого нужно выполнить следующие команды.
```bash
docker rm mysql-post
docker volume prune
```

После чего поднить контейнеры.
```bash
docker-compose up
```

И запустить репликацию.
```bash
./start_replication.sh
```

Для тестирования использовал запрос получения новостей пользователя, которые были реализованы в предыдцщем домашнем задании.
Создал в таблице news_line 1,5 млн записей.

Замеры до применения tarantool
```bash
wrk -t1 -c1000 -d60s --latency --script="config.lua" http://127.0.0.1:8881/post/api/v1/news/list
Running 1m test @ http://127.0.0.1:8881/post/api/v1/news/list
  1 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.26s   273.69ms   2.00s    78.60%
    Req/Sec   721.57    320.66     2.04k    68.81%
  Latency Distribution
     50%    1.23s 
     75%    1.41s 
     90%    1.59s 
     99%    1.92s 
  38497 requests in 1.00m, 3.89GB read
  Socket errors: connect 0, read 0, write 0, timeout 2194
  Non-2xx or 3xx responses: 161
Requests/sec:    640.48
Transfer/sec:     66.35MB
```

Замеры после
```bash
wrk -t1 -c1000 -d60s --latency --script="config.lua" http://127.0.0.1:8881/post/api/v1/news/list
Running 1m test @ http://127.0.0.1:8881/post/api/v1/news/list
  1 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.03s   223.26ms   2.00s    87.61%
    Req/Sec   811.87    448.30     3.52k    68.20%
  Latency Distribution
     50%    1.05s 
     75%    1.17s 
     90%    1.23s 
     99%    1.81s 
  44459 requests in 1.00m, 4.42GB read
  Socket errors: connect 0, read 0, write 0, timeout 5575
  Non-2xx or 3xx responses: 682
Requests/sec:    740.68
Transfer/sec:     75.41MB
```

Прирост производительности есть, но не могу сказать что существенный.
Сязано это с тем, что запрос простой и индексы MySQL работают хорошо.
# Отчёт по домашнему заданию №7

## Описание

Для того, чтобы можно было мастабировать обработку событий создания новых постов в данной работе был применён Routing Key.
Каждый пользователь имеет свой уникальный идентификатор. Routing Key пользователя вычисляется следующим образом - вычисляется сумма всей байт uid-а и берётся остаток от деления этой суммы на количество воркеров.
Из образа микросервиса post поднимается два обработчика. Сам микросервис post не читает сообщения из очереди, а только их туда пишет и обрабатывает пользовательский трафик.
Воркеры не обрабатывают пользовательский трафик, а занимаются только обработкой собщений из очереди.
Каждый воркер обрабатывает очередь по своему Routing Key. Такми образом получается что у каждого воркера своя очередь.
Также для повышения доступности RabbitMQ можно создать кластер rabbit.
А для повышения пропускной способности можно вместо использования Routing Key поднять несколько независимых экземпляров RabbitMQ.
При попадании на страницу с новостями "New Posts" пользователь видит текущие посты и подключается к realtime сервису для получения новых новостей через WebSocket.
Для получения новостей в реальном времени использован сервис centrifugo. Он поднят в двух экземплярах.
Пользователь подключается к одному realtime сервису, к какому определяется также по уникальному идентификатору пользователя как и с Routing Key - остаток от деления на количество realtime сервисов.
Ворек, который обрабатывает сообщения из RabbitMQ публикует пользователю новые сообщения в один realtime сервис, как оперделяется в какой описано выше.
Для передачи сообщений через центр сообщений для каждого пользователя используется свой канал. Например, post:05d8c592-0fb8-11ec-878a-aee2dbaf8f54.

Для корректно работы приложения нужно удалить старую базу микросервиса post. Для этого нужно выполнить следующие команды.
```bash
docker rm mysql-post
docker volume rm highloadarch_socialnetwork-db-post-data
```

После чего поднить контейнеры.
```bash
docker-compose up
```

И запустить репликацию.
```bash
./start_replication.sh
```

Для проверки работоспособности необходимо зарегистрировать двух новых пользователей. Зайти от первого пользователя и добавить в друзья второго.
После этого первым пользователем перейти на стриницу "New Posts".
Далее зайти от второго пользователя в другом браузере и создать новый пост на странице "Posts".
У первого пользователя появится новый пост.
# Отчёт по домашнему заданию №11

## Описание

Все микросервисы уже были развёрнуты в docker, поэтому в docker переносить ничего не пришлось.
Тажк consul уже использовался для работы vitess. Для балансировки нагрузки http запросов добавил fabio и fabio-grpc для grpc.
При поднятии микросервис регистрируется в consul, а из consul fabio узнаёт о микросервисах.
Bff socialnetwork взаимодействует с fabio, а fabio проксирует на нужный экземпляр.

Поднимаем контейнеры.
```bash
docker-compose up
```

И запустить репликацию.
```bash
./start_replication.sh
```

по адресу http://127.0.0.1:8500/ можно посмотреть состояние consul
по адресу http://127.0.0.1:9998/ и http://127.0.0.1:9918/ состояние fabio и fabio-grpc соответственно

Можно увидеть, что экземпляр conversation-1 успешно зарегистрировался и на него идёт трафик
```bash
conversation-1                      | http: 2022/06/29 20:40:28 map[args:conversationID:"2ed9932c-f70d-11ec-ac16-0242ac1e001e" duration:1.450719ms method:ListMessages requestID:b617a96b-f7eb-11ec-ab0c-0242ac1e0022] call finished
```

Далее поднимаем ещё 2 экземпляра conversation командой
```bash
docker-compose --file docker-compose-consul.yml up
```

В consul и fabio можно увидеть, что новые экземпляры зарегистрировались, а также можно видеть, что на них идёт трафик
```bash
conversation-3    | http: 2022/06/29 20:40:28 map[args:user:"fe2796fa-e4e6-11ec-9493-0242ac12000c"  target:"0eae91c0-ee7a-11ec-9067-0242ac12001e" duration:2.307091ms method:StartConversation requestID:b600c788-f7eb-11ec-ae69-0242ac1e0021] call finished
conversation-2    | http: 2022/06/29 20:40:28 map[args:conversationID:"2ed9932c-f70d-11ec-ac16-0242ac1e001e"  messages:"e7d4b7e4-f719-11ec-9039-0242ac1e001e"  messages:"8d896aa1-f71f-11ec-8e2f-0242ac1e001e"  messages:"8e21806c-f71f-11ec-984e-0242ac1e0020"  messages:"8e83683d-f71f-11ec-8e2f-0242ac1e001e"  messages:"8ee62627-f71f-11ec-984e-0242ac1e0020"  messages:"a3592406-f71f-11ec-8e2f-0242ac1e001e"  messages:"a3a868c1-f71f-11ec-984e-0242ac1e0020"  messages:"a3ec18e9-f71f-11ec-8e2f-0242ac1e001e"  messages:"a4845d07-f71f-11ec-984e-0242ac1e0020"  messages:"af1f98cc-f71f-11ec-8e2f-0242ac1e001e"  messages:"af6e966b-f71f-11ec-984e-0242ac1e0020"  messages:"afd30057-f71f-11ec-8e2f-0242ac1e001e"  messages:"b02a0f7f-f71f-11ec-984e-0242ac1e0020"  messages:"cf995aab-f71f-11ec-8e2f-0242ac1e001e"  messages:"d18394b1-f71f-11ec-984e-0242ac1e0020"  messages:"d34567da-f71f-11ec-8e2f-0242ac1e001e"  messages:"d5171298-f71f-11ec-984e-0242ac1e0020"  messages:"d49a4791-f7e6-11ec-9573-0242ac1e001f" duration:7.209799ms method:ReadMessages requestID:b600c788-f7eb-11ec-ae69-0242ac1e0021] call finished
```
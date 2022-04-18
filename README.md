# Учебный проект по курсу Highload Architect

Для запуска сервиса нужно выполнить команду:

```bash
docker-compose up
```

Для сборки и запуска сервиса нужно выполнить команды:

1) Дать права на запуск bash файла

```bash
chmod +x generate_data.sh
```

2) Запустить сборку
```bash
./build.sh
```

3) Поднять контейнеты
```bash
docker-compose up
```

Зайти в базу

```bash
docker exec -it mysql-node-1 mysql -usocialnetwork -ppasswd socialnetwork
```

Для генерации тестовых данных выполнить команду

```bash
chmod +x generate_data.sh
./generate_data.sh
```

Сервис доступен по url http://127.0.0.1:8881/app
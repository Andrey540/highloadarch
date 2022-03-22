# Учебный проект по курсу Highload Architect

Для запуска сервиса нужно выполнить команду:

```bash
docker-compose up
```

Для сборки и запуска сервиса нужно выполнить команды:

```bash
make
docker-compose up --build
docker-compose down
```

Зайти в базу

```bash
docker exec -it mysql mysql -usocialnetwork -ppasswd socialnetwork
```

Сервис доступен по url https://127.0.0.1:8884/
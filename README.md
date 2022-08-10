# Учебный проект по курсу Highload Architect

Для запуска приложения нужно выполнить команду:
```bash
./start.sh
```

Последующие команды нужно выполнить один раз при первом старте приложения.

Запустить автоматическую настройку забикса и дашборда графаны.
```bash
chmod +x ./zabbix/scripts/setup-zabbix.sh
./zabbix/scripts/setup-zabbix.sh
```

Запустить автоматическое создание дашборда графаны для прометея.
```bash
chmod +x ./prometheus/scripts/setup-prometheus.sh
./prometheus/scripts/setup-prometheus.sh
```

Сервис доступен по url http://127.0.0.1:8881/app

Для отсановки приложения нужно выполнить команду:
```bash
./stop.sh
```
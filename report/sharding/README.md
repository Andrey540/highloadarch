# Отчёт по домашнему заданию №4

## Запуск контейнера

Для того, чтобы поднять контейнеры нужно выполнить команду
```bash
docker-compose up
```
А зетем запустить репликацию с мастера на слейвы командой
```bash
chmod +x start_replication.sh
./start_replication.sh
```

Далее в ui создать у пользователей несколько бесед и написать сообщения

Схемы диалогов имеет следующий вид:

Таюлица, связывающая пользователя с другим пользователем (диалог). При создании диалога в таблицу записываются две записи
Отличаются они полями user_id и target. Поле user_id - кто ведёт переписку, поле target - с кем.
Собственно для каждгого пользователя будет создана запись, так что пользователь легко сможет найти все свои переписки.
Задача сделать так, чтобы переписки пользователя находились на одном шарде. Для этого будет использоват ключ шардирования по полю user_id.
```sql
CREATE TABLE IF NOT EXISTS user_conversation
(
    id INT AUTO_INCREMENT NOT NULL PRIMARY KEY,
    user_id binary(16) NOT NULL,
    conversation_id binary(16) NOT NULL,
    target binary(16) NOT NULL,
    INDEX `user_conversation_idx` (`user_id`, `conversation_id`),
    INDEX `conversation_idx` (`conversation_id`)
    ) ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
;
```

Таблица сообщений, в которой указано какой пользователь оставил сообщение и в какой переписке (поле conversation_id)
Для данной таблицы будет использован ключ шардирования conversation_id, чтобы можно было быстро вытащить все сообщения по переписке
```sql
CREATE TABLE IF NOT EXISTS message
(
    id              binary(16) NOT NULL PRIMARY KEY,
    user_id         binary(16) NOT NULL,
    conversation_id binary(16) NOT NULL,
    text            MEDIUMTEXT        DEFAULT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX           `conversation_idx` (`conversation_id`)
    ) ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
;
```

Для шардирования используется кластер vitess со схемой индексов.
```bash
{
	"sharded": true,
	"tables": {
		"conversation": {
			"column_vindexes": [
				{
					"column": "id",
					"name": "binary_md5_vdx"
				}
			]
		},
		"message": {
			"column_vindexes": [
				{
					"column": "conversation_id",
					"name": "binary_md5_vdx"
				}
			]
		},
		"user_conversation": {
			"column_vindexes": [
				{
					"column": "user_id",
					"name": "binary_md5_vdx"
				}
			]
		}
	},
	"vindexes": {
		"binary_md5_vdx": {
			"type": "binary_md5"
		},
		"hash": {
			"type": "hash"
		}
	}
}
```

В кластере создано 1 шард и у него есть реплика.

После того, как создано несколько переписок, можно зайти на шард и посмотреть что где находится.
Все последующие скрипты находятся в папке vitess
```bash
./vitess/dbcli.sh 101 vt_conversation_keyspace
```

Выполнить запрос
```sql
SELECT * FROM user_conversation;
```

Результат соответственно шардам
101
```sql
+----+------------------+------------------+------------------+
| id | user_id          | conversation_id  | target           |
+----+------------------+------------------+------------------+
|  1 | ��N�E쮦B�  | �P//�E��B�  | %U�S���&B�  |
|  2 | %U�S���&B�  | �P//�E��B�  | ��N�E쮦B�  |
|  3 | ��N�E쮦B�  | �i�p�E��B�  | t���&B�  |
|  4 | t���&B�  | �i�p�E��B�  | ��N�E쮦B�  |
|  5 | ��N�E쮦B�  | ��4i�E��B�  | 8e$���&B�  |
|  6 | 8e$���&B�  | ��4i�E��B�  | ��N�E쮦B�  |
|  7 | ��N�E쮦B�  | �.�$�E��B�  | jY��� �RB�  |
|  8 | jY��� �RB�  | �.�$�E��B�  | ��N�E쮦B�  |
|  9 | ��N�E쮦B�  | �����E��B�  | ,�W����&B�  |
| 10 | ,�W����&B�  | �����E��B�  | ��N�E쮦B�  |
| 11 | ��N�E쮦B�  | Ï���E��B�  | "�p	���&B�  |
| 12 | "�p	���&B�  | Ï���E��B�  | ��N�E쮦B�  |
+----+------------------+------------------+------------------+
```

Поднимает ещё два шарда 201 и 301
```bash
docker-compose --file docker-compose-shards.yml up
```

Запускаем решардинг
```bash
./vitess/reshard.sh
```

Переключаем чтение
```bash
./switch_reads.sh
```

Переключаем запись
```bash
./switch_writes.sh
```

После этого заходим на новые шарды 

```bash
./dbshardcli.sh 201 vt_conversation_keyspace
./dbshardcli.sh 301 vt_conversation_keyspace
```

Выполнить запрос
```sql
SELECT * FROM user_conversation;
```

201
```sql
+----+------------------+------------------+------------------+
| id | user_id          | conversation_id  | target           |
+----+------------------+------------------+------------------+
|  6 | 8e$���&B�  | ��4i�E��B�  | ��N�E쮦B�  |
|  8 | jY��� �RB�  | �.�$�E��B�  | ��N�E쮦B�  |
| 10 | ,�W����&B�  | �����E��B�  | ��N�E쮦B�  |
| 12 | "�p	���&B�  | Ï���E��B�  | ��N�E쮦B�  |
+----+------------------+------------------+------------------+
```

301
```sql
+----+------------------+------------------+------------------+
| id | user_id          | conversation_id  | target           |
+----+------------------+------------------+------------------+
|  1 | ��N�E쮦B�  | �P//�E��B�  | %U�S���&B�  |
|  2 | %U�S���&B�  | �P//�E��B�  | ��N�E쮦B�  |
|  3 | ��N�E쮦B�  | �i�p�E��B�  | t���&B�  |
|  4 | t���&B�  | �i�p�E��B�  | ��N�E쮦B�  |
|  5 | ��N�E쮦B�  | ��4i�E��B�  | 8e$���&B�  |
|  7 | ��N�E쮦B�  | �.�$�E��B�  | jY��� �RB�  |
|  9 | ��N�E쮦B�  | �����E��B�  | ,�W����&B�  |
| 11 | ��N�E쮦B�  | Ï���E��B�  | "�p	���&B�  |
+----+------------------+------------------+------------------+
```

Убеждаемся что данные среплицировались и ничего не потерялось

Стопаем первый шард
```bash
docker stop highloadarch_vttablet101_1
```

Проверяем, что приложение работает
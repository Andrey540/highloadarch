CREATE TABLE user
(
    id       binary(16) PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL
) ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
;
CREATE TABLE user
(
    id         binary(16) PRIMARY KEY,
    username   VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(255)        NOT NULL,
    last_name  VARCHAR(255)        NOT NULL,
    age        INT                 NOT NULL,
    sex        TINYINT             NOT NULL,
    interests  MEDIUMTEXT DEFAULT NULL,
    city       VARCHAR(255)        NOT NULL,
    password   VARCHAR(255)        NOT NULL
) ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
;
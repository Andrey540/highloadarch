CREATE TABLE IF NOT EXISTS conversation
(
    id    binary(16) PRIMARY KEY,
    data  JSON
    ) ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
;

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

CREATE TABLE IF NOT EXISTS stored_event
(
    id     binary(16) PRIMARY KEY,
    status INT          NOT NULL,
    type   VARCHAR(255) NOT NULL,
    body   JSON         NOT NULL,
    INDEX  `status_idx` (`status`)
);

CREATE TABLE IF NOT EXISTS processed_event
(
    id binary(16) NOT NULL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS processed_command
(
    id binary(16) NOT NULL PRIMARY KEY
);


CREATE TABLE message
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
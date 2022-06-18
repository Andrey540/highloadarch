CREATE TABLE IF NOT EXISTS user_unread_message
(
    id              INT AUTO_INCREMENT NOT NULL PRIMARY KEY,
    user_id         binary(16) NOT NULL,
    conversation_id binary(16) NOT NULL,
    count           INT DEFAULT 0,
    UNIQUE KEY      `user_conversation_idx` (`user_id`, `conversation_id`)
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


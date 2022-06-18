CREATE TABLE user_unread_message
(
    id              INT AUTO_INCREMENT NOT NULL PRIMARY KEY,
    user_id         binary(16) NOT NULL,
    conversation_id binary(16) NOT NULL,
    message_id      binary(16) NOT NULL,
    INDEX           `user_conversation_idx` (`conversation_id`, `user_id`),
    UNIQUE KEY      `user_mesage_idx` (`user_id`, `message_id`)
    ) ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
;
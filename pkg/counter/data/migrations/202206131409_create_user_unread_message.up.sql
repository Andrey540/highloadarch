CREATE TABLE user_unread_message
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
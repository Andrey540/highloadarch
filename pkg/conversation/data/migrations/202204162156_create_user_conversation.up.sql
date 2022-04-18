CREATE TABLE user_conversation
(
    id INT AUTO_INCREMENT NOT NULL PRIMARY KEY,
    user_id binary(16) NOT NULL,
    conversation_id binary(16) NOT NULL,
    INDEX `user_conversation_idx` (`user_id`, `conversation_id`),
    INDEX `=conversation_idx` (`conversation_id`)
) ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
;
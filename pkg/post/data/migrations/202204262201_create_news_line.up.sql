CREATE TABLE news_line
(
    id         INT(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id    VARCHAR(36) NOT NULL,
    post_id    VARCHAR(36) NOT NULL,
    author_id  VARCHAR(36) NOT NULL,
    title      VARCHAR(255) NOT NULL,
    INDEX      `user_idx` (`user_id`),
    INDEX      `post_idx` (`post_id`)
) ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
;
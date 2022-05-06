CREATE TABLE news_line
(
    id         INT(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id    binary(16) NOT NULL,
    post_id    binary(16) NOT NULL,
    post_index INT(11) NOT NULL DEFAULT 0,
    INDEX      `user_post_idx` (`user_id`, `post_index`),
    INDEX      `post_idx` (`post_id`)
) ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
;
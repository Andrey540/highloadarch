CREATE TABLE user_friend
(
    id        INT(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id   binary(16) NOT NULL,
    friend_id binary(16) NOT NULL,
    UNIQUE KEY `user_friend` (`user_id`,`friend_id`)
) ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
;
CREATE TABLE post
(
    id         binary(16) PRIMARY KEY,
    author_id  binary(16) NOT NULL,
    title      VARCHAR(255) NOT NULL,
    text       MEDIUMTEXT   NOT NULL,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
;
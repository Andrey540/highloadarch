CREATE TABLE conversation
(
    id    binary(16) PRIMARY KEY,
    data  JSON
) ENGINE = InnoDB
    CHARACTER SET = utf8mb4
    COLLATE utf8mb4_unicode_ci
;
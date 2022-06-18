CREATE TABLE stored_event
(
    id     binary(16) PRIMARY KEY,
    status INT          NOT NULL,
    type   VARCHAR(255) NOT NULL,
    body   JSON         NOT NULL,
    INDEX  `status_idx` (`status`)
);
CREATE FUNCTION `uuid_from_bin`(bin BINARY(16)) RETURNS varchar(36) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci
    DETERMINISTIC
    SQL SECURITY INVOKER
RETURN
  LOWER(CONCAT_WS('-',
                  SUBSTR(HEX(bin),  1, 8),
                  SUBSTR(HEX(bin),  9, 4),
                  SUBSTR(HEX(bin),  13, 4),
                  SUBSTR(HEX(bin),  17, 4),
                  SUBSTR(HEX(bin), 21)
    ));
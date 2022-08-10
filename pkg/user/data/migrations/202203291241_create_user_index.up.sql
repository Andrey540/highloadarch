ALTER TABLE user
    ADD INDEX user_username_idx (username),
    ADD INDEX user_first_name_idx (first_name),
    ADD INDEX user_last_name_idx (last_name);
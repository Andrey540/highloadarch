CREATE TABLE user_friend (
  user_id binary(16) PRIMARY KEY,
  friend_id binary(16) NOT NULL,
  UNIQUE KEY `user_friend` (`user_id`,`friend_id`)
);
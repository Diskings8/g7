CREATE TABLE `global_player_index` (
                                       `id` bigint NOT NULL AUTO_INCREMENT,
                                       `uid` bigint NOT NULL,
                                       `user_id` bigint NOT NULL,
                                       `server_id` int NOT NULL,
                                       `nickname` varchar(64) NOT NULL,
                                       `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
                                       PRIMARY KEY (`id`),
                                       UNIQUE KEY `uk_uid` (`uid`),
                                       KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `server_list` (
                               `id` bigint NOT NULL AUTO_INCREMENT,
                               `server_id` int NOT NULL,
                               `name` varchar(64) NOT NULL,
                               `addr` varchar(128) NOT NULL,
                               `db_name` varchar(64) NOT NULL,
                               `status` tinyint NOT NULL DEFAULT '1',
                               `channel` int NOT NULL DEFAULT 0,
                               `group_id` int NOT NULL DEFAULT 0,
                               PRIMARY KEY (`id`),
                               UNIQUE KEY `server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO server_list (server_id, name, addr, db_name, status)
VALUES (1, '测试服1', 'http://127.0.0.1:8082', 'game_db_s1', 1);
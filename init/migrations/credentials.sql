DROP TABLE IF EXISTS `credentials`;
CREATE TABLE `credentials` (
                         `user_id` BIGINT NOT NULL,
                         `server_name` varchar(100) NOT NULL,
                         `login` varchar(100) NOT NULL,
                         `password` varchar(100) NOT NULL,
                          CONSTRAINT `pk_credential` PRIMARY KEY (`user_id`, `server_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

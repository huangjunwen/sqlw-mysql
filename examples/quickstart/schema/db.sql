CREATE TABLE `user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `female` tinyint(1) DEFAULT NULL,
  `birthday` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
);

CREATE TABLE `employee` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `employee_sn` char(32) NOT NULL,
  `user_id` int(11) NOT NULL,
  `superior_id` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `employee_sn` (`employee_sn`),
  UNIQUE KEY `user_id` (`user_id`),
  KEY `fk_superior` (`superior_id`),
  CONSTRAINT `fk_superior` FOREIGN KEY (`superior_id`) REFERENCES `employee` (`id`),
  CONSTRAINT `fk_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
);

-- DDL

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
  KEY `superior_id` (`superior_id`),
  CONSTRAINT `fk_superior` FOREIGN KEY (`superior_id`) REFERENCES `employee` (`id`),
  CONSTRAINT `fk_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
);

-- DML

INSERT INTO `user` (`id`, `name`, `female`, `birthday`) VALUES 
  (1, "Ada", true, NULL),
  (2, "Bob", false, "1980-01-01"),
  (3, "Charlie", NULL, "1992-02-02"),
  (4, "Darwin", false, NULL),
  (5, "Eva", true, NULL),
  (6, "Franky", false, NULL),
  (7, "Grace", true, NULL),
  (26, "Zombie", NULL, NULL);

INSERT INTO `employee` (`id`, `employee_sn`, `user_id`, `superior_id`) VALUES
  (1, "SN-A", 1, NULL),
  (2, "SN-B", 2, 1),
  (3, "SN-C", 3, NULL),
  (4, "SN-D", 4, 3),
  (5, "SN-E", 5, 3),
  (6, "SN-F", 6, 4),
  (7, "SN-G", 7, NULL);


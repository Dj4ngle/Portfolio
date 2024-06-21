DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
                         `id` int(11) AUTO_INCREMENT PRIMARY KEY,
                         `username` varchar(200) NOT NULL,
                         `password` varchar(200) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `users` (`username`, `password`) VALUES
('rvasily',	'de847752bf50ff0aae49e7fcf81d189ac72a7db8086664f6737dc77442b35ee7');
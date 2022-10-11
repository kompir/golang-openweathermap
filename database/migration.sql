DROP DATABASE IF EXISTS weather;
CREATE DATABASE weather;
USE weather;
CREATE TABLE `meteo_table`
(
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    city_name VARCHAR(255),
    main_temp DECIMAL(10,2),
    date DATETIME
);
INSERT INTO `meteo_table` (`id`, `city_name`, `main_temp`, `date`) VALUES
                                                                       (1, 'Rousse', '10.00', '2022-10-04 00:00:00'),
                                                                       (2, 'Rousse', '11.00', '2022-10-05 00:00:00'),
                                                                       (3, 'Rousse', '12.00', '2022-10-06 00:00:00'),
                                                                       (4, 'Rousse', '13.00', '2022-10-07 00:00:00'),
                                                                       (5, 'Rousse', '14.00', '2022-10-08 00:00:00'),
                                                                       (6, 'Rousse', '15.00', '2022-10-09 00:00:00');
SET GLOBAL sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''));
CREATE TABLE `Registros` (
	`ID` int NOT NULL AUTO_INCREMENT,
	`Dispositivo` varchar(20) NOT NULL,
	`Temperatura_Celcius` float NOT NULL,
	`Fecha` timestamp NOT NULL DEFAULT current_timestamp(),
	`Humedad` int NOT NULL,
	PRIMARY KEY (`ID`)
)
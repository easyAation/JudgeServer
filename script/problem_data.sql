CREATE TABLE IF NOT EXISTS `problem_data` (
  `id`  INT NOT NULL AUTO_INCREMENT COMMENT "primary key",
  `pid` INT NOT NULL COMMENT "problem id",
  `input_file` varchar(100) NOT NULL COMMENT "input file path",
  `output_file` varchar(100) NOT NULL COMMENT "output file path",
  `md5` VARCHAR(100) NOT NULL COMMENT "",
  `md5_trim_space` VARCHAR(100) NOT NULL COMMENT "",
  PRIMARY KEY (id),
  UNIQUE KEY (input_file),
  UNIQUE KEY (output_file)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;
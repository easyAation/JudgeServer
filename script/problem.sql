CREATE TABLE IF NOT EXISTS `problem` (
  `id` INT NOT NULL COMMENT 'problem id',
  `name` VARCHAR(256) NOT NULL COMMENT 'problem name',
  `author` VARCHAR(256) NOT NULL COMMENT 'author ID',
  `status` VARCHAR(100) DEFAULT NULL DEFAULT "open" COMMENT 'status: open, close',
  `difficulty` VARCHAR(100) NOT NULL DEFAULT "" COMMENT 'difficulty',
  `case_data_input` VARCHAR(1000) NOT NULL DEFAULT "" COMMENT 'problem case data input',
  `case_data_output` VARCHAR(100) DEFAULT NULL DEFAULT "" COMMENT 'problem case data output',
  `description` TEXT NOT NULL COMMENT 'problem description',
  `input_des` VARCHAR(1000) NOT NULL DEFAULT "" COMMENT 'input description',
  `output_des` VARCHAR(1000) NOT NULL DEFAULT "" COMMENT 'output description',
  `hint` VARCHAR(1000) NOT NULL  DEFAULT "" COMMENT 'problem hint',
  `time_limit` INT NOT NULL COMMENT 'time limit',
  `memory_limit` INT NOT NULL COMMENT 'memory limit',
  `author_code` VARCHAR(1000) DEFAULT "" COMMENT 'author code',
  `created_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
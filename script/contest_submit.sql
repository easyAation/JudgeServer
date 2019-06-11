CREATE TABLE IF NOT EXISTS `contest_submit` (
  `id`   INT NOT NULL AUTO_INCREMENT COMMENT 'primary key',
   `cid` INT NOT NULL COMMENT 'contest id',
  `uid`  VARCHAR(100) NOT NULL COMMENT 'user id',
  `pid`  INT NOT NULL COMMENT 'problem ID',
  `submit_id` VARCHAR(22) NOT NULL COMMENT 'submit ID',
  `result` VARCHAR(20) NOT NULL DEFAULT "waiting" COMMENT 'value: Accept, WrongAnswer, Time_limit, MemoryLimit,MemoryLimit,RuntimeError,SystemError, PresentationError, InternalError',
  `code` VARCHAR(2000) DEFAULT "" COMMENT 'submit code',
  `language` VARCHAR(20) NOT NULL COMMENT 'value: C, CPP, GO',
  `memory` INT NOT NULL DEFAULT 0 COMMENT 'Programs Use memory',
  `run_time` INT NOT NULL DEFAULT 0 COMMENT 'Programs run time',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`submit_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
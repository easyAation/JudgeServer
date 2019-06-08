CREATE TABLE IF NOT EXISTS `contest_problem` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT 'primary key',
    `cid` INT NOT NULL  COMMENT 'contest key',
    `pid` INT NOT NULL COMMENT 'problem id',
    `position` INT NOT NULL COMMENT 'position',
    `submit_count` INT NOT NULL DEFAULT 0 COMMENT 'submit count',
    `solve_count` INT NOT NULL DEFAULT 0 COMMENT 'submit count',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
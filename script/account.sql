CREATE TABLE  IF NOT EXISTS `account` (
  `id` CHAR(22) NOT NULL COMMENT 'account id',
  `name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT 'name',
  `auth` VARCHAR(100) NOT NULL COMMENT 'local auth',
   githup_addr` VARCHAR(100) DEFAULT '' COMMENT 'githup addr',
   blog_addr` VARCHAR(100) DEFAULT '' COMMENT 'blog addr',
   PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

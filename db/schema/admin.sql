CREATE TABLE IF NOT EXISTS `user` (
  `user_id`    INT          NOT NULL AUTO_INCREMENT,
  `name`       VARCHAR(255) NOT NULL,
  `last_name`  VARCHAR(255),
  `email`      VARCHAR(255) NOT NULL UNIQUE,
  `password`   VARCHAR(255) NOT NULL,
  `is_active`  TINYINT(1)   NOT NULL DEFAULT 0,
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME,
  PRIMARY KEY (`user_id`)
);

CREATE TABLE IF NOT EXISTS `groups` (
  `group_id`   INT          NOT NULL AUTO_INCREMENT,
  `name`       VARCHAR(255) NOT NULL UNIQUE,
  `parent_id`  INT,
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME,
  PRIMARY KEY (`group_id`),
  FOREIGN KEY (`parent_id`) REFERENCES `groups` (`group_id`) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS `users_groups` (
  `user_group_id` INT NOT NULL AUTO_INCREMENT,
  `user_id`       INT NOT NULL,
  `group_id`      INT NOT NULL,
  PRIMARY KEY (`user_group_id`),
  UNIQUE (`user_id`, `group_id`),
  FOREIGN KEY (`user_id`)  REFERENCES `user`   (`user_id`)  ON DELETE CASCADE,
  FOREIGN KEY (`group_id`) REFERENCES `groups` (`group_id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `access_level` (
  `access_level_id` INT          NOT NULL AUTO_INCREMENT,
  `name`            VARCHAR(255) NOT NULL UNIQUE,
  PRIMARY KEY (`access_level_id`)
);

CREATE TABLE IF NOT EXISTS `users_access_levels` (
  `user_access_level_id` INT NOT NULL AUTO_INCREMENT,
  `user_id`              INT NOT NULL,
  `access_level_id`      INT NOT NULL,
  PRIMARY KEY (`user_access_level_id`),
  UNIQUE (`user_id`, `access_level_id`),
  FOREIGN KEY (`user_id`)         REFERENCES `user`         (`user_id`)         ON DELETE CASCADE,
  FOREIGN KEY (`access_level_id`) REFERENCES `access_level` (`access_level_id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `groups_access_levels` (
  `group_access_level_id` INT NOT NULL AUTO_INCREMENT,
  `group_id`              INT NOT NULL,
  `access_level_id`       INT NOT NULL,
  PRIMARY KEY (`group_access_level_id`),
  UNIQUE (`group_id`, `access_level_id`),
  FOREIGN KEY (`group_id`)        REFERENCES `groups`       (`group_id`)        ON DELETE CASCADE,
  FOREIGN KEY (`access_level_id`) REFERENCES `access_level` (`access_level_id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `module` (
  `module_id` INT          NOT NULL AUTO_INCREMENT,
  `name`      VARCHAR(255) NOT NULL UNIQUE,
  PRIMARY KEY (`module_id`)
);

CREATE TABLE IF NOT EXISTS `access_levels_modules` (
  `access_level_module_id` INT          NOT NULL AUTO_INCREMENT,
  `access_level_id`        INT          NOT NULL,
  `module_id`              INT          NOT NULL,
  `data_scope`             ENUM('OWN','RESPONSIBLE','GROUP','ALL') NOT NULL DEFAULT 'ALL',
  `permission`             INT          NOT NULL DEFAULT 0,
  `can_be_assigned`        TINYINT(1)   NOT NULL DEFAULT 0,
  `created_at`             DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`             DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at`             DATETIME,
  PRIMARY KEY (`access_level_module_id`),
  UNIQUE (`access_level_id`, `module_id`),
  FOREIGN KEY (`access_level_id`) REFERENCES `access_level` (`access_level_id`) ON DELETE CASCADE,
  FOREIGN KEY (`module_id`)       REFERENCES `module`        (`module_id`)       ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `menu` (
  `menu_id`        INT          NOT NULL AUTO_INCREMENT,
  `name`           VARCHAR(255) NOT NULL,
  `sort_index`     INT          NOT NULL,
  `paths`          VARCHAR(255) NOT NULL,
  `module_id`      INT          NOT NULL,
  `parent_id`      INT,
  `required_scope` ENUM('OWN','RESPONSIBLE','GROUP','ALL'),
  PRIMARY KEY (`menu_id`),
  FOREIGN KEY (`module_id`) REFERENCES `module` (`module_id`),
  FOREIGN KEY (`parent_id`) REFERENCES `menu`   (`menu_id`)
);

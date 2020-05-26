SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for metrics
-- ----------------------------
DROP TABLE IF EXISTS `metrics`;
CREATE TABLE `metrics` (
  `id` int(11) NOT NULL COMMENT '主键',
  `strategy_id` int(11) NOT NULL COMMENT '策略id',
  `metric` varchar(50) NOT NULL COMMENT 'metric名称',
  `value` float NOT NULL COMMENT '值',
  `step` int(11) NOT NULL DEFAULT '30' COMMENT '步长',
  `type` varchar(255) NOT NULL,
  `timestamp` int(11) NOT NULL COMMENT '时间戳',
  `tags` text COMMENT '标签',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for strategy
-- ----------------------------
DROP TABLE IF EXISTS `strategy`;
CREATE TABLE `strategy` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'pk',
  `name` varchar(100) NOT NULL COMMENT '任务名称',
  `note` varchar(255) DEFAULT NULL COMMENT '备注',
  `mode` int(11) NOT NULL DEFAULT '0' COMMENT '任务类型',
  `context` text NOT NULL COMMENT '任务详情',
  `is_delete` int(11) NOT NULL DEFAULT '0' COMMENT '是否删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

SET FOREIGN_KEY_CHECKS = 1;

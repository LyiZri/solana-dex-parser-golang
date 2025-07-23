-- 修复 most_earn_token_win_rate 字段精度问题
-- 
-- 问题: 当前 DECIMAL(6,4) 只能存储最大值 99.9999
-- 解决: 改为 DECIMAL(10,4) 可以存储最大值 999999.9999 (6位整数，4位小数)
-- 
-- 在加密货币投资中，1000倍收益虽然极端但是可能的，特别是早期 meme 币投资

-- 修改 user_report 表
ALTER TABLE `user_report` 
MODIFY COLUMN `most_earn_token_win_rate` DECIMAL(10,4) DEFAULT NULL COMMENT '最盈利代币的盈利倍数，支持最大999999.9999倍收益';

-- 如果存在 smart_season_one 表，也需要修改
ALTER TABLE `smart_season_one` 
MODIFY COLUMN `most_earn_token_win_rate` DECIMAL(10,4) DEFAULT NULL COMMENT '最盈利代币的盈利倍数，支持最大999999.9999倍收益';

-- 验证修改结果
DESCRIBE user_report;

-- 显示字段定义
SHOW CREATE TABLE user_report; 
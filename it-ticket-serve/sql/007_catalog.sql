USE it_ticket;

CREATE TABLE IF NOT EXISTS ticket_templates (
  id          BIGINT PRIMARY KEY AUTO_INCREMENT,
  name        VARCHAR(32)  NOT NULL,
  category    ENUM('hardware','software','network','other') NOT NULL,
  title_hint  VARCHAR(120) NOT NULL,
  hint        VARCHAR(80)  NOT NULL DEFAULT '',
  icon        VARCHAR(16)  NOT NULL DEFAULT 'other',
  sort_order  INT          NOT NULL DEFAULT 0,
  enabled     TINYINT(1)   NOT NULL DEFAULT 1,
  fields      JSON         NOT NULL,
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_templates_enabled_sort (enabled, sort_order, id)
);

CREATE TABLE IF NOT EXISTS canned_replies (
  id          BIGINT PRIMARY KEY AUTO_INCREMENT,
  title       VARCHAR(80)   NOT NULL,
  body        VARCHAR(2000) NOT NULL,
  sort_order  INT           NOT NULL DEFAULT 0,
  enabled     TINYINT(1)    NOT NULL DEFAULT 1,
  created_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_replies_enabled_sort (enabled, sort_order, id)
);

INSERT INTO ticket_templates (name, category, title_hint, hint, icon, sort_order, enabled, fields)
SELECT name, category, title_hint, hint, icon, sort_order, enabled, fields
FROM (
  SELECT
    '打印机' AS name,
    'hardware' AS category,
    '打印机无法使用' AS title_hint,
    '卡纸、不进纸、驱动问题' AS hint,
    'printer' AS icon,
    10 AS sort_order,
    1 AS enabled,
    '[{"key":"model","label":"型号","placeholder":"例如 HP LaserJet M227","required":true},{"key":"location","label":"位置","placeholder":"楼层、房间或工位","required":true},{"key":"symptom","label":"现象","placeholder":"卡纸、不进纸、驱动报错等","required":true},{"key":"tried","label":"已试步骤","placeholder":"重启、重装驱动、换过纸没有","required":true}]' AS fields
  UNION ALL
  SELECT
    '邮箱',
    'software',
    '邮箱登不上',
    '登不上、收不到信',
    'email',
    20,
    1,
    '[{"key":"account","label":"邮箱账号","placeholder":"name@company.com","required":true},{"key":"client","label":"客户端","placeholder":"网页 / Outlook / 手机","required":true},{"key":"error","label":"报错原文","placeholder":"把提示原样贴上来","required":true},{"key":"tried","label":"已试步骤","placeholder":"改过密码、清过缓存没有","required":true}]'
  UNION ALL
  SELECT
    '网络',
    'network',
    '连不上网',
    '有线或无线连不上',
    'network',
    30,
    1,
    '[{"key":"location","label":"地点","placeholder":"工位、会议室或楼层","required":true},{"key":"kind","label":"接入方式","placeholder":"有线 / 无线","required":true},{"key":"symptom","label":"现象","placeholder":"完全不通、很慢、只能上内网","required":true},{"key":"tried","label":"已试步骤","placeholder":"换过网线、重连 Wi-Fi、重启没有","required":true}]'
) seed
WHERE NOT EXISTS (SELECT 1 FROM ticket_templates LIMIT 1);

INSERT INTO canned_replies (title, body, sort_order, enabled)
SELECT title, body, sort_order, enabled
FROM (
  SELECT '请补充报错原文' AS title, '请把完整报错原文贴上来（不要只写「打不开」），并说明大概从什么时候开始。' AS body, 10 AS sort_order, 1 AS enabled
  UNION ALL SELECT '请先重启再试', '请先重启设备后再试一次，把重启后是否恢复、以及新的报错原文回在这条评论下。', 20, 1
  UNION ALL SELECT '已远程处理，请确认', '已远程处理，请确认是否恢复。如已恢复请关闭工单；还有问题请直接在下面补充。', 30, 1
  UNION ALL SELECT '已更换配件，请确认', '配件已更换，请确认设备是否恢复正常。正常的话请关闭工单。', 40, 1
  UNION ALL SELECT '需要现场处理', '这条需要现场处理。请回复方便上门的时间段和具体位置。', 50, 1
  UNION ALL SELECT '账号已重置', '账号已重置，请查收新的登录方式并尽快改密。登录成功后请关闭工单。', 60, 1
) seed
WHERE NOT EXISTS (SELECT 1 FROM canned_replies LIMIT 1);

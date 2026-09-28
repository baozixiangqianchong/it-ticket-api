-- 账号类通知没有工单。ticket_id 改为可空。
-- 已有「标题=事件、正文=工单标题」的旧数据改成「标题=#编号 工单标题」。
USE it_ticket;

ALTER TABLE notifications
  MODIFY COLUMN ticket_id BIGINT NULL;

UPDATE notifications n
INNER JOIN tickets t ON t.id = n.ticket_id
SET
  n.body = IF(n.body IS NULL OR n.body = '' OR n.body = t.title, n.title, n.body),
  n.title = LEFT(CONCAT('#', n.ticket_id, ' ', t.title), 120)
WHERE n.ticket_id IS NOT NULL
  AND n.title NOT LIKE '#%';

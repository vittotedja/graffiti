-- Drop indexes first
DROP INDEX IF EXISTS idx_notifications_recipient_id;
DROP INDEX IF EXISTS idx_notifications_created_at;

-- Then drop the table
DROP TABLE IF EXISTS notifications;

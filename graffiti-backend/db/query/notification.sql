-- name: CreateNotification :one
INSERT INTO notifications (
  recipient_id,
  sender_id,
  type,
  entity_id,
  message,
  is_read,
  created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetNotificationsByUser :many
SELECT * FROM notifications
WHERE recipient_id = $1
ORDER BY created_at DESC;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications
WHERE recipient_id = $1 AND is_read = false;

-- name: MarkNotificationAsRead :exec
UPDATE notifications
SET is_read = true
WHERE id = $1;

-- name: MarkAllNotificationsAsRead :exec
UPDATE notifications
SET is_read = true
WHERE recipient_id = $1;

-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = $1;
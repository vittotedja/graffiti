export type NotificationType = 
  | 'friend_request'
  | 'friend_request_accepted'
  | 'post_like'
  | 'wall_post';

export interface Notification {
  id: string;
  recipient_id: string;
  sender_id: string;
  type: NotificationType;
  entity_id: string;
  message: string;
  is_read: boolean;
  created_at: string;
  // Additional fields after joining with user data
  sender_name?: string;
  sender_username?: string;
  sender_avatar?: string;
}

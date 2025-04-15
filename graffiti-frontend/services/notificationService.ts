import { Notification } from '@/types/notification';
import { fetchWithAuth } from '@/lib/auth';

export const notificationService = {
  // Get all notifications for the current user
  getNotifications: async (): Promise<Notification[]> => {
    const response = await fetchWithAuth('/api/v1/notifications', {
      method: 'GET',
    });
    
    if (!response.ok) {
      throw new Error('Failed to fetch notifications');
    }
    
    return response.json();
  },

  // Mark a notification as read
  markAsRead: async (id: string): Promise<void> => {
    const response = await fetchWithAuth(`/api/v1/notifications/${id}/read`, {
      method: 'PUT',
    });
    
    if (!response.ok) {
      throw new Error('Failed to mark notification as read');
    }
  },

  // Mark all notifications as read
  markAllAsRead: async (): Promise<void> => {
    const response = await fetchWithAuth('/api/v1/notifications/read-all', {
      method: 'PUT',
    });
    
    if (!response.ok) {
      throw new Error('Failed to mark all notifications as read');
    }
  },
  
  // Get unread notification count
  getUnreadCount: async (): Promise<number> => {
    const response = await fetchWithAuth('/api/v1/notifications/unread/count', {
      method: 'GET',
    });
    
    if (!response.ok) {
      throw new Error('Failed to get unread notification count');
    }
    
    const data = await response.json();
    return data.count;
  }
};

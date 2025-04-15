"use client";
import { useEffect, useState } from 'react';
import { Bell } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { notificationService } from '@/services/notificationService';
import Link from 'next/link';
import { useUser } from '@/hooks/useUser';

export function NotificationBadge() {
  const { user, loading: userLoading } = useUser();
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Only fetch notifications if user is logged in
    if (!user || userLoading) return;

    const fetchUnreadCount = async () => {
      try {
        setLoading(true);
        const count = await notificationService.getUnreadCount();
        setUnreadCount(count);
      } catch (error) {
        console.error('Failed to fetch unread count:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchUnreadCount();

    // Set up polling to check for new notifications every minute
    const intervalId = setInterval(fetchUnreadCount, 60000);
    
    return () => clearInterval(intervalId);
  }, [user, userLoading]);

  // Don't render anything if user is not logged in or still loading
  if (!user || userLoading) return null;

  return (
    <Link href="/notifications">
      <Button variant="ghost" size="icon" className="relative">
        <Bell className="h-5 w-5" />
        {!loading && unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 bg-destructive text-primary-foreground text-xs rounded-full h-5 w-5 flex items-center justify-center">
            {unreadCount > 99 ? '99+' : unreadCount}
          </span>
        )}
      </Button>
    </Link>
  );
}

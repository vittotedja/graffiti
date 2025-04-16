"use client";
import { useEffect, useState, useCallback } from 'react';
import { Bell } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { notificationService } from '@/services/notificationService';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useUser } from '@/hooks/useUser';

export function NotificationBadge() {
    const { user, loading: userLoading } = useUser();
    const [unreadCount, setUnreadCount] = useState(0);
    const [loading, setLoading] = useState(true);
    const pathname = usePathname();
    
    // Function to fetch unread count
    const fetchUnreadCount = useCallback(async () => {
        if (!user || userLoading) return;
        
        try {
            setLoading(true);
            const count = await notificationService.getUnreadCount();
            setUnreadCount(count);
        } catch (error) {
            console.error('Failed to fetch unread count:', error);
        } finally {
            setLoading(false);
        }
    }, [user, userLoading]);
  
    useEffect(() => {
        fetchUnreadCount();
        
        // Set up polling to check for new notifications every minute
        const intervalId = setInterval(fetchUnreadCount, 60000);
        
        return () => clearInterval(intervalId);
    }, [fetchUnreadCount]);
    
    // Reset unread count when visiting the notifications page
    useEffect(() => {
        if (pathname === '/notifications') {
            setUnreadCount(0);
        }
    }, [pathname]);
  
    if (!user || userLoading) return null;
  
    return (
        <Link href="/notifications">
            <Button variant="ghost" size="icon" className="relative" aria-label="Notifications">
            <Bell className="h-5 w-5" />
            {!loading && unreadCount > 0 && (
                <span className="absolute -top-1 -right-1 flex items-center justify-center h-5 w-5 text-[10px] font-medium rounded-full bg-destructive text-destructive-foreground">
                {unreadCount > 99 ? '99+' : unreadCount}
                </span>
            )}
            </Button>
        </Link>
    );
}
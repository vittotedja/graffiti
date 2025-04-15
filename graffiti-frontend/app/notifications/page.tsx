"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import {
	ChevronLeft,
	UserPlus,
	Heart,
	MessageSquare,
	ImageIcon,
	Check,
	X,
	Loader2,
} from "lucide-react";

import {Button} from "@/components/ui/button";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Card, CardContent} from "@/components/ui/card";
import { notificationService } from "@/services/notificationService";
import { useUser } from "@/hooks/useUser";
import { formatDistanceToNow, set } from "date-fns";
import { useRouter } from "next/navigation";
import { fetchWithAuth } from "@/lib/auth";
import { toast } from "sonner";
import { Notification } from "@/types/notification";

export default function NotificationsPage() {
	const { user, loading: userLoading } = useUser(true); // Redirect if not logged in
  	const [notifications, setNotifications] = useState<Notification[]>([]);
  	const [loading, setLoading] = useState(true);
  	const router = useRouter();
	const [localDate, setLocalDate] = useState(new Date());
	
	  useEffect(() => {
		// Only fetch notifications if user is logged in
		if (!user || userLoading) return;
		
		const fetchNotifications = async () => {
		  try {
			setLoading(true);
			const data = await notificationService.getNotifications();
			
			// Enhance notifications with sender details from the user object if available
			const enhancedNotifications = await Promise.all(
			  data.map(async (notification) => {
				// If the sender is the current user, we already have the details
				if (notification.sender_id === user.id) {
					setLocalDate(new Date(notification.created_at));
				  return {
					...notification,
					sender_name: user.fullname,
					sender_username: user.username,
					sender_avatar: user.profile_picture,
				  };
				}
				
				// Otherwise, we need to fetch the sender details
				try {
				  const response = await fetchWithAuth(`/api/v1/users/${notification.sender_id}`);
				  if (response.ok) {
					const sender = await response.json();
					setLocalDate(new Date(notification.created_at));
					return {
					  ...notification,
					  sender_name: sender.fullname,
					  sender_username: sender.username,
					  sender_avatar: sender.profile_picture,
					};
				  }
				} catch (error) {
				  console.error(`Failed to fetch sender details for notification ${notification.id}:`, error);
				}
				
				// Return the original notification if we couldn't fetch sender details
				return notification;
			  })
			);
			
			setNotifications(enhancedNotifications);
		  } catch (error) {
			console.error("Failed to fetch notifications:", error);
			toast.error("Failed to load notifications");
		  } finally {
			setLoading(false);
		  }
		};
	
		fetchNotifications();
	  }, [user, userLoading]);

	// Function to render notification icon based on type
	const getNotificationIcon = (type: string) => {
		switch (type) {
		  case "wall_post":
			return <ImageIcon className="h-4 w-4" />;
		  case "friend_request":
			return <UserPlus className="h-4 w-4" />;
		  case "friend_request_accepted":
			return <UserPlus className="h-4 w-4" />;
		  case "post_like":
			return <Heart className="h-4 w-4" />;
		  default:
			return <ImageIcon className="h-4 w-4" />;
		}
	};

	if (userLoading || loading) {
		return (
		  <div className="flex justify-center items-center min-h-screen">
			<Loader2 className="h-8 w-8 animate-spin text-primary" />
		  </div>
		);
	}

	if (!user) return null; // This shouldn't happen due to the redirect in useUser

	return (
		<div className="min-h-screen bg-[url('/images/concrete-texture.jpg')] bg-cover">
			<div className="container mx-auto px-4 pb-20">
				{/* Header */}
				<header className="py-4 flex items-center gap-2">
					<Link href="/">
						<Button variant="ghost" size="icon" className="rounded-full">
							<ChevronLeft className="h-6 w-6" />
						</Button>
					</Link>
					<h1 className="text-2xl md:text-3xl font-bold font-graffiti">
						Notifications
					</h1>
				</header>

				{/* Notifications List */}
				<div className="space-y-3 mt-4">
					{notifications.map((notification) => (
						<Card
							key={notification.id}
							className={`border-2 ${
								notification.is_read ? "border-primary/10" : "border-primary/30"
							} bg-black/5 backdrop-blur-sm`}
						>
							<CardContent className="p-4">
								<div className="flex items-center gap-3">
									<div
										className={`p-2 rounded-full ${
											notification.is_read ? "bg-muted" : "bg-primary/20"
										}`}
									>
										{getNotificationIcon(notification.type)}
									</div>
									<Link href={`/profile/${notification.sender_id}`}>
										<Avatar>
											<AvatarImage
												src={notification.sender_avatar}
												alt={notification.sender_name}
											/>
											<AvatarFallback>
												{notification.sender_name ? notification.sender_name.charAt(0) : "U"}
											</AvatarFallback>
									</Avatar>
									</Link>
									<div className="flex-1">
										<div className="flex items-center gap-1">
											{/* <span className="font-medium">
											</span> */}
											<span>@{notification.message ? notification.message : "Message not available."}</span>
										</div>
										{/* <div className="text-xs text-muted-foreground">
											{formatDistanceToNow(new Date(notification.created_at), { addSuffix: true })}
										</div> */}
									</div>

									{notification.type === "friend_request" && !notification.is_read && (
										<div className="flex items-center gap-2">
											<Button size="sm" className="h-8 w-8 p-0 rounded-full">
												<Check className="h-4 w-4" />
											</Button>
											<Button
												size="sm"
												variant="outline"
												className="h-8 w-8 p-0 rounded-full"
											>
												<X className="h-4 w-4" />
											</Button>
										</div>
									)}

									{notification.type === "friend_request_accepted" && (
										<Link href={`/profile/${notification.sender_id}`}>
											<Button size="sm" variant="outline">
												View Profile
											</Button>
										</Link>
									)}
								</div>
							</CardContent>
						</Card>
					))}
				</div>
			</div>
		</div>
	);
}

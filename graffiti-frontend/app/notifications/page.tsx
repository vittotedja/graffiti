import Link from "next/link";
import {
	ChevronLeft,
	UserPlus,
	Heart,
	MessageSquare,
	ImageIcon,
	Check,
	X,
} from "lucide-react";

import {Button} from "@/components/ui/button";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Card, CardContent} from "@/components/ui/card";
import {MobileNav} from "@/components/mobile-nav";

export default function NotificationsPage() {
	// Mock data for notifications
	const notifications = [
		{
			id: 1,
			type: "post",
			user: {
				name: "Friend Name",
				username: "friendname",
				avatar: "/placeholder.svg?height=40&width=40",
			},
			content: 'posted on your "Birthday Wall"',
			time: "2h ago",
			read: false,
		},
		{
			id: 2,
			type: "friend_request",
			user: {
				name: "Friend Name 2",
				username: "friendname2",
				avatar: "/placeholder.svg?height=40&width=40",
			},
			content: "wants to befriend you!",
			time: "5h ago",
			read: false,
		},
		{
			id: 3,
			type: "friend_request",
			user: {
				name: "Friend Name 3",
				username: "friendname3",
				avatar: "/placeholder.svg?height=40&width=40",
			},
			content: "wants to befriend you!",
			time: "1d ago",
			read: false,
		},
		{
			id: 4,
			type: "like",
			user: {
				name: "Friend Name 4",
				username: "friendname4",
				avatar: "/placeholder.svg?height=40&width=40",
			},
			content: "liked your post on Other Friend Name's wall",
			time: "2d ago",
			read: true,
		},
		{
			id: 5,
			type: "friend_accept",
			user: {
				name: "Friend Name 5",
				username: "friendname5",
				avatar: "/placeholder.svg?height=40&width=40",
			},
			content: "accepts your friend request",
			time: "3d ago",
			read: true,
		},
	];

	// Function to render notification icon based on type
	const getNotificationIcon = (type: string) => {
		switch (type) {
			case "post":
				return <ImageIcon className="h-4 w-4" />;
			case "friend_request":
			case "friend_accept":
				return <UserPlus className="h-4 w-4" />;
			case "like":
				return <Heart className="h-4 w-4" />;
			case "comment":
				return <MessageSquare className="h-4 w-4" />;
			default:
				return <ImageIcon className="h-4 w-4" />;
		}
	};

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
								notification.read ? "border-primary/10" : "border-primary/30"
							} bg-black/5 backdrop-blur-sm`}
						>
							<CardContent className="p-4">
								<div className="flex items-center gap-3">
									<div
										className={`p-2 rounded-full ${
											notification.read ? "bg-muted" : "bg-primary/20"
										}`}
									>
										{getNotificationIcon(notification.type)}
									</div>
									<Avatar>
										<AvatarImage
											src={notification.user.avatar}
											alt={notification.user.name}
										/>
										<AvatarFallback>
											{notification.user.name.charAt(0)}
										</AvatarFallback>
									</Avatar>
									<div className="flex-1">
										<div className="flex items-center gap-1">
											<span className="font-medium">
												@{notification.user.username}
											</span>
											<span>{notification.content}</span>
										</div>
										<div className="text-xs text-muted-foreground">
											{notification.time}
										</div>
									</div>

									{notification.type === "friend_request" && (
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

									{notification.type === "friend_accept" && (
										<Button size="sm" variant="outline">
											View Profile
										</Button>
									)}
								</div>
							</CardContent>
						</Card>
					))}
				</div>
			</div>

			{/* Mobile Navigation */}
			<MobileNav />
		</div>
	);
}

"use client";

import {useState} from "react";
import Link from "next/link";
import {UserWithMutualFriends} from "@/types/user";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardFooter} from "@/components/ui/card";
import {Badge} from "@/components/ui/badge";
import {ClockIcon, UserPlus, Users} from "lucide-react";
import {formatFullName} from "@/lib/formatter";
import Image from "next/image";
import {fetchWithAuth} from "@/lib/auth";
import {toast} from "sonner";

interface UserCardProps {
	user: UserWithMutualFriends;
	onViewMutualFriends: () => void;
	friendshipStatus?: "none" | "pending" | "friends";
}

export function UserCard({
	user,
	onViewMutualFriends,
	friendshipStatus = "none",
}: UserCardProps) {
	const [requestSent, setRequestSent] = useState(
		friendshipStatus === "pending"
	);
	const [isLoading, setIsLoading] = useState(false);

	async function sendFriendRequest(userId: string): Promise<void> {
		try {
			const response = await fetchWithAuth("/api/v1/friend-requests", {
				method: "POST",
				body: JSON.stringify({
					to_user_id: userId,
				}),
			});
			if (!response.ok) throw new Error("Error Adding Friends");
			toast.success("Successfully sent friend request!");
		} catch (error) {
			console.error(error);
			toast.warning("Error Adding Friends");
		}
	}

	const handleSendRequest = async () => {
		setIsLoading(true);
		try {
			await sendFriendRequest(user.id);
			setRequestSent(true);
		} catch (error) {
			console.error("Failed to send friend request:", error);
		} finally {
			setIsLoading(false);
		}
	};

	return (
		<>
			<Card className="overflow-hidden">
				<Image
					src={user.background_image || "/mockbg.jpg"}
					width={1000}
					height={400}
					alt="User Background"
				/>
				<CardContent className="p-6 pt-0 -mt-12">
					<div className="flex flex-col items-center">
						<Avatar className="h-24 w-24 border-4 border-background">
							<AvatarImage src={user.profile_picture} alt={user.fullname} />
							<AvatarFallback>{formatFullName(user.fullname)}</AvatarFallback>
						</Avatar>
						<div className="mt-4 text-center">
							<Link href={`/profile/${user.id}`} className="hover:underline">
								<h3 className="font-semibold text-lg">{user.fullname}</h3>
							</Link>
							<p className="text-sm text-muted-foreground">@{user.username}</p>
						</div>

						<div className="mt-4 flex items-center gap-1.5">
							<Badge
								variant="secondary"
								className="cursor-pointer"
								onClick={onViewMutualFriends}
							>
								<Users className="h-3 w-3 mr-1" />
								{user.mutual_friend_count} mutual friend(s)
							</Badge>
						</div>
					</div>
				</CardContent>
				<CardFooter className="flex justify-center p-6 pt-0">
					<Button
						className="w-full"
						onClick={handleSendRequest}
						disabled={requestSent || isLoading}
					>
						{requestSent ? (
							<>
								<ClockIcon className="h-4 w-4 mr-2" />
								Request Pending
							</>
						) : (
							<>
								<UserPlus className="h-4 w-4 mr-2" />
								Add Friend
							</>
						)}
					</Button>
				</CardFooter>
			</Card>
		</>
	);
}

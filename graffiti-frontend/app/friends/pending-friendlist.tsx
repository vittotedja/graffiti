"use client";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Button} from "@/components/ui/button";
import {fetchWithAuth} from "@/lib/auth";
import {formatFullName} from "@/lib/formatter";
import {Friendship} from "@/types/friends";
import {Ban, Check} from "lucide-react";
import Link from "next/link";
import {useEffect, useState} from "react";
import {toast} from "sonner";

export default function PendingFriendsList() {
	const [pendingFriends, setPendingFriends] = useState<Friendship[]>([]);

	const fetchPendingFriends = async () => {
		try {
			const response = await fetchWithAuth(`/api/v1/friends?type=requested`, {
				method: "GET",
			});
			if (!response.ok) {
				throw new Error("Failed to fetch pending friends");
			}
			const data = await response.json();
			setPendingFriends(data);
		} catch (error) {
			console.error("Error fetching pending friends:", error);
		}
	};

	useEffect(() => {
		fetchPendingFriends();
	}, []);

	const acceptFriends = async (friendship_id: string) => {
		if (!friendship_id) toast.error("no friends selected");
		try {
			const response = await fetchWithAuth(`/api/v1/friend-requests/accept`, {
				method: "PUT",
				body: JSON.stringify({
					friendship_id: friendship_id,
				}),
			});
			if (!response.ok) {
				throw new Error("Failed to accept friends");
			}
			const friend = pendingFriends.find((f) => f.ID === friendship_id);
			const name = friend?.Fullname || friendship_id;
			toast.success(`You are now friends with ${name}`);
			await fetchPendingFriends();
		} catch (error) {
			console.error("Error fetching pending friends:", error);
			toast.error("Failed to accept friend request");
		}
	};

	if (!pendingFriends || !pendingFriends.length)
		return <div className="h-8 text-center">No pending friend requests</div>;

	return (
		<div className="divide-y">
			{pendingFriends &&
				pendingFriends.map((friend) => (
					<div
						key={friend.UserID}
						className="flex items-center justify-between p-4 hover:bg-accent/50"
					>
						<Link
							href={`/profile/${friend.UserID}`}
							className="flex items-center gap-3 hover:underline cursor-pointer"
						>
							<Avatar>
								<AvatarImage
									src={friend.ProfilePicture}
									alt={friend.Fullname}
								/>
								<AvatarFallback>
									{formatFullName(friend.Fullname)}
								</AvatarFallback>
							</Avatar>
							<div>
								<div className="font-medium">{friend.Fullname}</div>
								<div className="text-xs text-muted-foreground">
									@{friend.Username}
								</div>
							</div>
						</Link>
						<div className="flex items-center gap-2">
							<Button
								size="sm"
								onClick={() => {
									acceptFriends(friend.ID);
								}}
							>
								<Check className="h-4 w-4 mr-2" />
								Accept
							</Button>
							<Button
								variant="outline"
								size="sm"
								onClick={() => {
									acceptFriends(friend.ID);
								}}
							>
								<Ban className="h-4 w-4 mr-2" />
								Reject
							</Button>
						</div>
					</div>
				))}
		</div>
	);
}

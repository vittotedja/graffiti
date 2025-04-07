"use client";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Button} from "@/components/ui/button";
import {fetchWithAuth} from "@/lib/auth";
import {formatFullName} from "@/lib/formatter";
import {Friendship} from "@/types/friends";
import {X} from "lucide-react";
import {useEffect, useState} from "react";

export default function RequestedFriendsList() {
	const [requestedFriends, setRequestedFriends] = useState<Friendship[]>([]);

	useEffect(() => {
		const fetchPendingFriends = async () => {
			try {
				const response = await fetchWithAuth(`/api/v1/friends?type=sent`, {
					method: "GET",
				});
				if (!response.ok) {
					throw new Error("Failed to fetch pending friends");
				}
				const data = await response.json();
				setRequestedFriends(data);
			} catch (error) {
				console.error("Error fetching pending friends:", error);
			}
		};
		fetchPendingFriends();
	}, []);

	if (!requestedFriends || !requestedFriends.length)
		return <div className="h-8 text-center">No sent friend requests</div>;

	return (
		<div className="divide-y">
			{requestedFriends &&
				requestedFriends.map((friend) => (
					<div
						key={friend.UserID}
						className="flex items-center justify-between p-4 hover:bg-accent/50"
					>
						<div className="flex items-center gap-3">
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
						</div>

						<Button size="sm" variant={"destructive"}>
							<X className="h-4 w-4 mr-2" />
							Cancel
						</Button>
					</div>
				))}
		</div>
	);
}

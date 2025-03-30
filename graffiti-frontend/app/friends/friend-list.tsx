"use client";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Button} from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {fetchWithAuth} from "@/lib/auth";
import {formatFullName} from "@/lib/formatter";
import {Friendship} from "@/types/friends";
import {MoreVertical} from "lucide-react";
import {useEffect, useState} from "react";

export default function FriendsList() {
	const [pendingFriends, setPendingFriends] = useState<Friendship[]>([]);

	useEffect(() => {
		const fetchPendingFriends = async () => {
			try {
				const response = await fetchWithAuth(
					`http://localhost:8080/api/v1/friends?type=friends`,
					{
						method: "GET",
					}
				);
				if (!response.ok) {
					throw new Error("Failed to fetch pending friends");
				}
				const data = await response.json();
				setPendingFriends(data);
			} catch (error) {
				console.error("Error fetching pending friends:", error);
			}
		};
		fetchPendingFriends();
	}, []);

	if (!pendingFriends || !pendingFriends.length)
		return <div className="h-8 text-center">No friends yet</div>;

	return (
		<div className="divide-y">
			{pendingFriends &&
				pendingFriends.map((friend) => (
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
						<DropdownMenu>
							<DropdownMenuTrigger asChild>
								<Button variant="ghost" size="icon" className="h-8 w-8">
									<MoreVertical className="h-4 w-4" />
								</Button>
							</DropdownMenuTrigger>
							<DropdownMenuContent align="end">
								<DropdownMenuItem>View Profile</DropdownMenuItem>
								<DropdownMenuSeparator />
								<DropdownMenuItem>Block</DropdownMenuItem>
								<DropdownMenuItem className="text-destructive">
									Remove Friend
								</DropdownMenuItem>
							</DropdownMenuContent>
						</DropdownMenu>
					</div>
				))}
		</div>
	);
}

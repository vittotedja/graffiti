"use client";

import {useState, useEffect} from "react";
import {X} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card";
// import {fetchDiscoverFriendsByMutuals} from "@/lib/api";
import {UserWithMutualFriends} from "@/types/user";
import {fetchWithAuth} from "@/lib/auth";
import {formatFullName} from "@/lib/formatter";

interface RecommendedFriendsProps {
	onClose: () => void;
	friendId: string;
}

export default function RecommendedFriends({
	onClose,
	friendId,
}: RecommendedFriendsProps) {
	const [recommendedFriends, setRecommendedFriends] = useState<
		UserWithMutualFriends[]
	>([]);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		async function fetchDiscoverFriendsByMutuals(): Promise<
			UserWithMutualFriends[]
		> {
			try {
				const response = await fetchWithAuth("/api/v1/friends/discover", {
					method: "POST",
				});
				if (!response.ok) throw new Error("Something went wrong");
				const data = await response.json();
				if (!data) return [];
				const filteredData = data.filter(
					(user: UserWithMutualFriends) => user.id !== friendId
				);
				console.log("unfiltered", data);
				console.log("filtered", filteredData);
				return filteredData;
			} catch (error) {
				console.error(error);
			}
			return [];
		}

		const loadRecommendedFriends = async () => {
			try {
				const data = await fetchDiscoverFriendsByMutuals();
				// Only show top 3 recommended friends
				setRecommendedFriends(data.slice(0, 3));
			} catch (error) {
				console.error("Failed to load recommended friends:", error);
			} finally {
				setLoading(false);
			}
		};

		loadRecommendedFriends();
	}, [friendId]);

	return (
		<Card className="w-full mb-8 relative">
			<Button
				variant="ghost"
				size="icon"
				className="absolute right-2 top-2"
				onClick={onClose}
			>
				<X className="h-4 w-4" />
			</Button>
			<CardHeader className="py-4">
				<CardTitle className="text-lg">People you might know</CardTitle>
			</CardHeader>
			<CardContent>
				{loading ? (
					<div className="flex gap-4 overflow-x-auto py-2">
						{[...Array(3)].map((_, i) => (
							<div key={i} className="flex-shrink-0 w-[200px]">
								<div className="h-[100px] rounded-lg bg-muted animate-pulse" />
							</div>
						))}
					</div>
				) : (
					<div className="flex gap-4 overflow-x-auto py-2">
						{recommendedFriends.length <= 0 && (
							<div className="text-center flex text-sm mx-auto mb-5 items-center">
								No recommended friends yet
							</div>
						)}
						{recommendedFriends.map((friend) => (
							<div
								key={friend.id}
								className="flex-shrink-0 w-[200px] border rounded-lg p-3"
							>
								<div className="flex items-center gap-3 mb-2">
									<Avatar>
										<AvatarImage
											src={friend.profile_picture}
											alt={friend.fullname}
										/>
										<AvatarFallback>
											{formatFullName(friend.fullname)}
										</AvatarFallback>
									</Avatar>
									<div>
										<p className="font-medium text-sm line-clamp-1">
											{friend.fullname}
										</p>
										<p className="text-xs text-muted-foreground">
											@{friend.username}
										</p>
									</div>
								</div>
								<div className="text-xs text-muted-foreground mb-3">
									{friend.mutual_friend_count} mutual friends
								</div>
								<Button size="sm" className="w-full text-xs">
									Add Friend
								</Button>
							</div>
						))}
					</div>
				)}
			</CardContent>
		</Card>
	);
}

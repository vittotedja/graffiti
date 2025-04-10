"use client";

import {useState, useEffect} from "react";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogDescription,
} from "@/components/ui/dialog";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {User} from "@/types/user";
import Link from "next/link";
import {formatFullName} from "@/lib/formatter";
import {fetchWithAuth} from "@/lib/auth";
import {useUser} from "@/hooks/useUser";

interface MutualFriendsDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	userId: string;
	userName: string;
}

export function MutualFriendsDialog({
	open,
	onOpenChange,
	userId,
	userName,
}: MutualFriendsDialogProps) {
	const {user} = useUser();
	const [mutualFriends, setMutualFriends] = useState<User[]>([]);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		async function fetchMutualFriends(userId: string): Promise<User[]> {
			if (!user) return [];
			try {
				const response = await fetchWithAuth("/api/v1/friends/mutual", {
					method: "POST",
					body: JSON.stringify({
						user_id: userId,
					}),
				});
				if (!response.ok) throw new Error("Failed to fetch");
				const data = await response.json();
				console.log(data);
				return data;
			} catch (e) {
				console.error(e);
			}

			return [];
		}

		if (open) {
			const loadMutualFriends = async () => {
				setLoading(true);
				try {
					const data = await fetchMutualFriends(userId);
					setMutualFriends(data);
				} catch (error) {
					console.error("Failed to load mutual friends:", error);
				} finally {
					setLoading(false);
				}
			};

			loadMutualFriends();
		}
	}, [open, userId, user]);

	// Add guard to prevent rendering if user name is missing
	if (!userName) {
		// Close the dialog if it's open but userName is missing
		if (open) onOpenChange(false);
		return null;
	}

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-md">
				<DialogHeader>
					<DialogTitle>Mutual Friends with {userName}</DialogTitle>
					<DialogDescription>
						You and {userName} have {mutualFriends.length} mutual friends
					</DialogDescription>
				</DialogHeader>

				<div className="max-h-[60vh] overflow-y-auto py-4">
					{loading ? (
						<div className="space-y-4">
							{[...Array(3)].map((_, i) => (
								<div key={i} className="flex items-center gap-4">
									<div className="h-10 w-10 rounded-full bg-muted animate-pulse" />
									<div className="space-y-2">
										<div className="h-4 w-32 bg-muted animate-pulse rounded" />
										<div className="h-3 w-24 bg-muted animate-pulse rounded" />
									</div>
								</div>
							))}
						</div>
					) : mutualFriends.length === 0 ? (
						<p className="text-center py-8 text-muted-foreground">
							No mutual friends found
						</p>
					) : (
						<div className="space-y-4">
							{mutualFriends.map((friend) => (
								<div key={friend.id} className="flex items-center gap-4">
									<Avatar>
										<AvatarImage
											src={friend.profile_picture}
											alt={friend.fullname}
										/>
										<AvatarFallback>
											{friend.fullname ? formatFullName(friend.fullname) : "NA"}
										</AvatarFallback>
									</Avatar>
									<div>
										<Link
											href={`/profile/${friend.id}`}
											className="font-medium hover:underline"
										>
											{friend.fullname}
										</Link>
										<p className="text-sm text-muted-foreground">
											@{friend.username}
										</p>
									</div>
								</div>
							))}
						</div>
					)}
				</div>
			</DialogContent>
		</Dialog>
	);
}

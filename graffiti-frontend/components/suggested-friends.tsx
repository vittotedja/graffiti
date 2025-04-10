"use client";

import {useState, useEffect, useMemo} from "react";
import {UserCard} from "./user-card";
import {MutualFriendsDialog} from "./mutual-friends-dialog";
import {UserWithMutualFriends} from "@/types/user";
import {fetchWithAuth} from "@/lib/auth";

// Simplified friendship status type
type FriendshipStatus = {
	userId: string;
	status: "none" | "pending" | "friends";
};

export default function SuggestedFriends() {
	const [allSuggestedFriends, setAllSuggestedFriends] = useState<
		UserWithMutualFriends[]
	>([]);
	const [friendshipStatuses, setFriendshipStatuses] = useState<
		FriendshipStatus[]
	>([]);
	const [loading, setLoading] = useState(true);
	const [selectedUser, setSelectedUser] =
		useState<UserWithMutualFriends | null>(null);
	const [showMutualFriends, setShowMutualFriends] = useState(false);

	// Filter the suggested friends to exclude those who are already friends
	// This is computed once when dependencies change, not on every render
	const filteredFriends = useMemo(() => {
		const statusMap: Record<string, string> = {};

		// Create a lookup map for faster access
		friendshipStatuses.forEach((status) => {
			statusMap[status.userId] = status.status;
		});

		// Filter out users who are already friends
		return allSuggestedFriends.filter(
			(user) => statusMap[user.id] !== "friends"
		);
	}, [allSuggestedFriends, friendshipStatuses]);

	async function fetchDiscoverFriendsByMutuals(): Promise<
		UserWithMutualFriends[]
	> {
		try {
			const response = await fetchWithAuth("/api/v1/friends/discover", {
				method: "POST",
			});
			if (!response.ok) throw new Error("Something went wrong");
			const data = await response.json();
			return data;
		} catch (error) {
			console.error(error);
		}
		return [];
	}

	async function fetchFriendshipStatuses(
		userIds: string[]
	): Promise<FriendshipStatus[]> {
		try {
			// Create an array of promises for each friendship status check
			const statusPromises = userIds.map(async (userId) => {
				try {
					const response = await fetchWithAuth("/api/v1/friendships", {
						method: "POST",
						body: JSON.stringify({
							to_user_id: userId,
						}),
					});

					if (!response.ok) {
						return {userId, status: "none" as const};
					}

					const data = await response.json();
					const statusString = data?.Status?.Status;

					// Type the status explicitly
					const status: "none" | "pending" | "friends" =
						statusString === "pending"
							? "pending"
							: statusString === "friends"
							? "friends"
							: "none";

					return {userId, status};
				} catch (error) {
					console.error(
						"Error fetching friendship status for user",
						userId,
						error
					);
					return {userId, status: "none" as const};
				}
			});

			return Promise.all(statusPromises);
		} catch (error) {
			console.error("Failed to fetch friendship statuses:", error);
			return userIds.map((id) => ({userId: id, status: "none" as const}));
		}
	}

	useEffect(() => {
		const loadSuggestedFriends = async () => {
			try {
				// Fetch all suggested friends first
				const data = await fetchDiscoverFriendsByMutuals();
				if (!data) return;
				setAllSuggestedFriends(data);

				// Fetch all friendship statuses in one go
				if (data.length > 0) {
					const userIds = data.map((user) => user.id);
					const statuses = await fetchFriendshipStatuses(userIds);
					setFriendshipStatuses(statuses);
				}
			} catch (error) {
				console.error("Failed to load suggested friends:", error);
			} finally {
				setLoading(false);
			}
		};

		loadSuggestedFriends();
	}, []);

	const handleViewMutualFriends = (user: UserWithMutualFriends) => {
		if (user && user.id && user.fullname) {
			setSelectedUser(user);
			setShowMutualFriends(true);
		}
	};

	const handleCloseDialog = (open: boolean) => {
		setShowMutualFriends(open);
		if (!open) {
			// Reset selectedUser when dialog closes
			setTimeout(() => setSelectedUser(null), 300);
		}
	};

	if (loading) {
		return (
			<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
				{[...Array(6)].map((_, i) => (
					<div
						key={i}
						className="h-[280px] rounded-lg bg-muted animate-pulse"
					/>
				))}
			</div>
		);
	}

	// No need to filter in the render - we're using the pre-filtered list
	return (
		<>
			{filteredFriends.length <= 0 && (
				<div className="text-center py-12">
					<h3 className="text-lg font-medium">No recommended friends yet</h3>
					<p className="text-muted-foreground mt-2">
						We&apos;re working on recommending friends for you! In the meantime,
						expand your network all on your own!
					</p>
				</div>
			)}
			<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
				{filteredFriends.map((user) => {
					// Find the friendship status for this user
					const friendStatus =
						friendshipStatuses.find((fs) => fs.userId === user.id)?.status ||
						"none";

					return (
						<UserCard
							key={user.id}
							user={user}
							friendshipStatus={friendStatus}
							onViewMutualFriends={() => handleViewMutualFriends(user)}
						/>
					);
				})}
			</div>
			{selectedUser && selectedUser.id && selectedUser.fullname && (
				<MutualFriendsDialog
					open={showMutualFriends}
					onOpenChange={handleCloseDialog}
					userId={selectedUser.id}
					userName={selectedUser.fullname}
				/>
			)}
		</>
	);
}

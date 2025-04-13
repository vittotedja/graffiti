"use client";

import {useState, useEffect} from "react";
import {UserCard} from "./user-card";
import {MutualFriendsDialog} from "./mutual-friends-dialog";
import {UserWithMutualFriends} from "@/types/user";
// import { fetchTrendingUsers } from "@/lib/api"

export default function TrendingUsers() {
	const [trendingUsers, setTrendingUsers] = useState<UserWithMutualFriends[]>(
		[]
	);
	const [loading, setLoading] = useState(true);
	const [selectedUser, setSelectedUser] =
		useState<UserWithMutualFriends | null>(null);
	const [showMutualFriends, setShowMutualFriends] = useState(false);

	async function fetchTrendingUsers(): Promise<UserWithMutualFriends[]> {
		return [];
	}

	useEffect(() => {
		const loadTrendingUsers = async () => {
			try {
				const data = await fetchTrendingUsers();
				setTrendingUsers(data);
			} catch (error) {
				console.error("Failed to load trending users:", error);
			} finally {
				setLoading(false);
			}
		};

		loadTrendingUsers();
	}, []);

	const handleViewMutualFriends = (user: UserWithMutualFriends) => {
		setSelectedUser(user);
		setShowMutualFriends(true);
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

	return (
		<>
			{trendingUsers.length <= 0 && (
				<div className="text-center py-12">
					<h3 className="text-lg font-medium">Explore trending users soon!</h3>
					<p className="text-muted-foreground mt-2">
						We&apos;re working on recommending friends for you! In the meantime,
						expand your network all on your own!
					</p>
				</div>
			)}
			<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
				{trendingUsers.map((user) => (
					<UserCard
						key={user.id}
						user={user}
						onViewMutualFriends={() => handleViewMutualFriends(user)}
					/>
				))}
			</div>

			{selectedUser && (
				<MutualFriendsDialog
					open={showMutualFriends}
					onOpenChange={setShowMutualFriends}
					userId={selectedUser.id}
					userName={selectedUser.fullname}
				/>
			)}
		</>
	);
}

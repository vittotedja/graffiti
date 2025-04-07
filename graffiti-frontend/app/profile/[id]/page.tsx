"use client";

import {useEffect, useState} from "react";
import Image from "next/image";
import {WallGrid} from "@/components/wall-grid";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {formatFullName} from "@/lib/formatter";

import {usePathname} from "next/navigation";
import {fetchWithAuth} from "@/lib/auth";
import {User} from "@/types/user";
import {Button} from "@/components/ui/button";
import {ClockIcon, UserMinus2, UserPlus2} from "lucide-react";
import {toast} from "sonner";
import {useUser} from "@/hooks/useUser";

export default function ProfilePage() {
	const {user: loggedInUser} = useUser();
	const pathname = usePathname();
	const splitPath = pathname.split("/");
	const [user, setUser] = useState<User>();
	const [friendshipID, setFriendshipID] = useState<string>("");
	const [friendshipStatus, setFriendshipStatus] = useState<
		"friends" | "pending" | false
	>(false);
	const [friendshipFromUserID, setFriendshipFromUserID] = useState<
		string | null
	>(null);

	const userId = splitPath[2];

	useEffect(() => {
		const fetchUser = async () => {
			if (!userId) return;
			try {
				const response = await fetchWithAuth(`/api/v1/users/${userId}`);
				if (!response) return;
				const data = await response.json();
				setUser(data);

				const isFriendResp = await fetchWithAuth("/api/v1/friendships", {
					method: "POST",
					body: JSON.stringify({
						to_user_id: userId,
					}),
				});
				if (!isFriendResp.ok) return;
				const isFriendsData = await isFriendResp.json();
				if (isFriendsData) {
					setFriendshipID(isFriendsData.ID);

					const status = isFriendsData.Status.Status;
					const fromUser = isFriendsData.FromUser;

					if (status === "friends") {
						setFriendshipStatus("friends");
					} else if (status === "pending") {
						setFriendshipStatus("pending");
						setFriendshipFromUserID(fromUser);
					} else {
						setFriendshipStatus(false);
						setFriendshipFromUserID(null);
					}
				} else {
					setFriendshipStatus(false);
					setFriendshipFromUserID(null);
				}
			} catch (err) {
				console.error("Failed to fetch wall data:", err);
			}
		};

		fetchUser();
	}, [userId]);

	if (splitPath[1] != "profile") return;

	const addFriend = async () => {
		if (!user) return;
		try {
			const response = await fetchWithAuth("/api/v1/friend-requests", {
				method: "POST",
				body: JSON.stringify({
					to_user_id: user.id,
				}),
			});
			if (!response.ok) throw new Error("Error Adding Friends");
			toast.success("Successfully sent friend request!");
		} catch (error) {
			console.error(error);
			toast.warning("Error Adding Friends");
		}
	};

	const removeFriend = async () => {
		if (!user) return;
		console.log(friendshipID);
		try {
			const response = await fetchWithAuth("/api/v1/friendships", {
				method: "DELETE",
				body: JSON.stringify({
					friendship_id: friendshipID,
				}),
			});
			if (!response.ok) throw new Error("Error Removing Friends");
			toast.success("Successfully removed friend!");
		} catch (error) {
			console.error(error);
			toast.warning("Error Removing Friends");
		}
	};

	const acceptFriend = async (friendship_id: string) => {
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
			toast.success(`You are now friends with ${user?.fullname}`);
			setFriendshipStatus("friends");
		} catch (error) {
			console.error("Error fetching pending friends:", error);
			toast.error("Failed to accept friend request");
		}
	};

	if (!user) return <p>Not logged in</p>;

	return (
		<div className="min-h-screen bg-[url('/images/concrete-texture.jpg')] bg-cover">
			<div className="container mx-auto px-4 pb-20">
				{/* Home Content */}
				<main className="mt-6">
					<div className="relative mb-8 rounded-xl overflow-hidden">
						{/* Background image with more visibility */}
						<div className="absolute inset-0 bg-gradient-to-b from-transparent via-transparent to-black/60"></div>
						<Image
							src={user.background_image || "/mockbg.jpg"}
							alt="Home Banner"
							width={1200}
							height={400}
							className="w-full h-[250px] md:h-[350px] object-cover"
						/>

						{/* Bottom section with avatar and buttons */}
						<div className="absolute bottom-0 left-0 right-0 p-4 md:p-6 flex items-end justify-between">
							{/* Avatar on bottom left */}
							<div className="flex items-end gap-4">
								<Avatar className="h-16 w-16 md:h-24 md:w-24 border-4 border-background">
									<AvatarImage
										src={
											user.profile_picture ||
											"/placeholder.svg?height=96&width=96"
										}
										alt="User Avatar"
									/>
									<AvatarFallback>
										{formatFullName(user.fullname)}
									</AvatarFallback>
								</Avatar>
								<div className="mb-1 md:mb-2">
									<h2 className="text-2xl md:text-3xl font-bold text-white font-graffiti">
										{user.fullname}
									</h2>
									<h2 className="text-md md:text-md font-medium text-white/55 font-graffiti">
										@{user.username}
									</h2>
									<h2 className="text-sm italic text-white/55 font-graffiti">
										{user.bio}
									</h2>
								</div>
							</div>
							<div className="flex gap-2 md:gap-3">
								{friendshipStatus === "friends" && (
									<Button
										variant="destructive"
										onClick={removeFriend}
										className="text-xs md:text-sm h-8 md:h-9"
									>
										<UserMinus2 />
										Remove Friend
									</Button>
								)}

								{friendshipStatus === "pending" &&
									friendshipFromUserID === loggedInUser?.id && (
										<Button
											variant="outline"
											disabled
											className="text-xs md:text-sm h-8 md:h-9"
										>
											<ClockIcon /> {/* pending clock icon */}
											Pending
										</Button>
									)}

								{friendshipStatus === "pending" &&
									friendshipFromUserID !== loggedInUser?.id && (
										<Button
											onClick={() => acceptFriend(friendshipID)}
											className="bg-green-600/50 hover:bg-green-600 text-xs md:text-sm h-8 md:h-9 text-white"
										>
											<UserPlus2 />
											Accept Friend
										</Button>
									)}

								{friendshipStatus === false && (
									<Button
										variant="outline"
										onClick={addFriend}
										className="bg-black/50 text-white border-white/20 hover:bg-black/70 hover:text-blue-500 text-xs md:text-sm h-8 md:h-9"
									>
										<UserPlus2 />
										Add Friend
									</Button>
								)}
							</div>
						</div>
					</div>
					{/* Walls */}
					<WallGrid userId={userId} />
				</main>
			</div>
		</div>
	);
}

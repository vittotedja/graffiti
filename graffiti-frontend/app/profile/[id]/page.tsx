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
import {UserMinus2, UserPlus2} from "lucide-react";
import {toast} from "sonner";

export default function ProfilePage() {
	const pathname = usePathname();
	const splitPath = pathname.split("/");
	const [user, setUser] = useState<User>();
	const [isFriend, setIsFriend] = useState<boolean>(false);

	const userId = splitPath[2];

	useEffect(() => {
		const fetchUser = async () => {
			if (!userId) return;
			try {
				const response = await fetchWithAuth(
					`http://localhost:8080/api/v1/users/${userId}`
				);
				if (!response) return;
				const data = await response.json();
				setUser(data);

				const isFriendResp = await fetchWithAuth(
					"http://localhost:8080/api/v1/friendships",
					{
						method: "POST",
						body: JSON.stringify({
							to_user_id: userId,
						}),
					}
				);
				const isFriendsData = await isFriendResp.json();
				console.log(isFriendsData.Status.Status);
				setIsFriend(isFriendsData.Status.Status == "friends");
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
			const response = await fetchWithAuth(
				"http://localhost:8080/api/v1/friend-requests",
				{
					method: "POST",
					body: JSON.stringify({
						to_user_id: user.id,
					}),
				}
			);
			if (!response.ok) throw new Error("Error Adding Friends");
			toast.success("Successfully sent friend request!");
		} catch (error) {
			console.error(error);
			toast.warning("Error Adding Friends");
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
								{!isFriend ? (
									<Button
										variant="outline"
										className="bg-black/50 text-white border-white/20 hover:bg-black/70 text-xs md:text-sm h-8 md:h-9"
										onClick={addFriend}
									>
										<UserPlus2 />
										Add Friend
									</Button>
								) : (
									<Button
										variant="destructive"
										className=" border-white/20 hover:bg-red-500/50 text-xs md:text-sm h-8 md:h-9"
										// onClick={addFriend}
									>
										<UserMinus2 />
										Remove Friend
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

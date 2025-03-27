"use client";

import {useEffect, useState} from "react";
import Image from "next/image";
import {Archive, Menu, Pencil, Plus, X} from "lucide-react";

import {Button} from "@/components/ui/button";
import {WallGrid} from "@/components/wall-grid";
import {CreateWallModal} from "@/components/create-wall-modal";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";

import {cn} from "@/lib/utils";
import {fetchWithAuth} from "@/lib/auth";
import {useUser} from "@/hooks/useUser";

export default function HomePage() {
	const {user, loading} = useUser();
	const [createWallModalOpen, setCreateWallModalOpen] = useState(false);
	const [fabOpen, setFabOpen] = useState(false);

	// Optional ripple effect when FAB is clicked
	useEffect(() => {
		if (fabOpen) {
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			const handleClickOutside = (e: any) => {
				const fabElement = document.querySelector(".fab-container");
				if (fabElement && !fabElement.contains(e.target)) {
					setFabOpen(false);
				}
			};

			document.addEventListener("mousedown", handleClickOutside);
			return () => {
				document.removeEventListener("mousedown", handleClickOutside);
			};
		}
	}, [fabOpen]);

	const toggleFab = () => {
		setFabOpen(!fabOpen);
	};

	const fetchWallData = async () => {
		try {
			const response = await fetchWithAuth(
				"http://localhost:8080/api/v1/walls"
			);
			if (!response) return; // already redirected if 401

			const data = await response.json();
			console.log("Wall data:", data);
		} catch (err) {
			console.error("Failed to fetch wall data:", err);
		}
	};

	useEffect(() => {
		fetchWallData();
	}, []);

	if (loading) return <p>Loading...</p>;
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
							src="/mockbg.jpg"
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
										src="/placeholder.svg?height=96&width=96"
										alt="User Avatar"
									/>
									<AvatarFallback>UN</AvatarFallback>
								</Avatar>
								<div className="mb-1 md:mb-2">
									<h2 className="text-2xl md:text-3xl font-bold text-white font-graffiti">
										@JohnDoe
									</h2>
								</div>
							</div>

							{/* Buttons on bottom right */}
							<div className="flex gap-2 md:gap-3">
								<Button
									variant="outline"
									className="bg-black/50 text-white border-white/20 hover:bg-black/70 text-xs md:text-sm h-8 md:h-9"
								>
									<Pencil />
									Edit Home
								</Button>
							</div>
						</div>
					</div>
					{/* Walls */}
					<WallGrid />
				</main>
			</div>
			<div
				className={cn(
					"fixed inset-0 bg-black/60 z-40 transition-opacity duration-300",
					fabOpen ? "opacity-100" : "opacity-0 pointer-events-none"
				)}
				onClick={() => setFabOpen(false)}
			/>

			{/* Floating Action Button with Radial Menu */}
			<div className="fixed bottom-6 right-6 z-50 fab-container">
				{/* Main FAB Button */}
				<Button
					onClick={toggleFab}
					variant={"special"}
					className="h-14 w-14 rounded-full shadow-lg flex items-center justify-center transition-all duration-300 z-50 relative"
				>
					<span className="relative z-10 transition-transform duration-300">
						{fabOpen ? (
							<X className="h-6 w-6 text-white transition-all duration-300" />
						) : (
							<Menu className="h-6 w-6 text-white transition-all duration-300" />
						)}
					</span>
				</Button>

				{/* Radial Menu Items */}
				<div className="absolute inset-0 z-40">
					{/* Expanding Circle Animation */}
					<div
						className={cn(
							"absolute inset-0 rounded-full bg-gray-700/20 transition-transform duration-300",
							fabOpen ? "scale-[4.5]" : "scale-0"
						)}
					></div>
					{/* Create New Wall Option */}
					<div
						className={cn(
							"absolute flex items-center gap-3 transition-all duration-500",
							fabOpen
								? "opacity-100 translate-x-[-200px] translate-y-[-100px]"
								: "opacity-0 translate-x-0 translate-y-0 pointer-events-none"
						)}
						style={{transitionDelay: fabOpen ? "0.1s" : "0s"}}
					>
						<div className="bg-white text-xs font-medium text-black px-3 py-1.5 rounded-full shadow-lg whitespace-nowrap">
							Create New Wall
						</div>
						<Button
							className="h-12 w-12 rounded-full flex items-center justify-center shadow-lg z-100 hover:scale-110"
							onClick={() => setCreateWallModalOpen(true)}
						>
							<Plus className="h-5 w-5" />
						</Button>
					</div>

					{/* View Archives Option */}
					<div
						className={cn(
							"absolute flex items-center gap-3 transition-all duration-500",
							fabOpen
								? "opacity-100 translate-x-[-225px] translate-y-[-40px]"
								: "opacity-0 translate-x-0 translate-y-0 pointer-events-none"
						)}
						style={{transitionDelay: fabOpen ? "0.2s" : "0s"}}
					>
						<div className="bg-white text-xs font-medium text-black px-3 py-1.5 rounded-full shadow-lg whitespace-nowrap">
							View Archives
						</div>
						<Button className="h-12 w-12 rounded-full bg-primary flex items-center justify-center shadow-lg z-100 hover:scale-110">
							<Archive className="h-5 w-5" />
						</Button>
					</div>
				</div>
			</div>
			{/* Create Wall Modal */}
			<CreateWallModal
				isOpen={createWallModalOpen}
				onClose={() => setCreateWallModalOpen(false)}
				onCreateWall={(data) => {
					console.log("Creating wall:", data);
					// Here you would typically call an API to create the wall
					setCreateWallModalOpen(false);
				}}
			/>
		</div>
	);
}

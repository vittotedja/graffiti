"use client";

import {useState} from "react";
import Image from "next/image";
import {Plus} from "lucide-react";

import {Button} from "@/components/ui/button";
import {WallGrid} from "@/components/wall-grid";
import {CreateWallModal} from "@/components/create-wall-modal";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";

export default function HomePage() {
	const [createWallModalOpen, setCreateWallModalOpen] = useState(false);
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
										Welcome to My Home
									</h2>
									<p className="text-white/80 text-sm md:text-base">
										Express yourself through digital graffiti
									</p>
								</div>
							</div>

							{/* Buttons on bottom right */}
							<div className="flex gap-2 md:gap-3">
								<Button
									className="bg-primary hover:bg-primary/90 text-white dark:text-black text-xs md:text-sm h-8 md:h-9"
									onClick={() => setCreateWallModalOpen(true)}
								>
									<Plus className="mr-1 md:mr-2 h-3 w-3 md:h-4 md:w-4" /> Create
									Wall
								</Button>
								<Button
									variant="outline"
									className="bg-black/50 text-white border-white/20 hover:bg-black/70 text-xs md:text-sm h-8 md:h-9"
								>
									Edit Home
								</Button>
							</div>
						</div>
					</div>
					{/* Walls */}
					<WallGrid />
				</main>
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

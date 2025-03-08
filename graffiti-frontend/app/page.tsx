"use client";

import {useState} from "react";
import Image from "next/image";
import {Plus} from "lucide-react";

import {Button} from "@/components/ui/button";
import {WallGrid} from "@/components/wall-grid";
import {CreateWallModal} from "@/components/create-wall-modal";

export default function HomePage() {
	const [createWallModalOpen, setCreateWallModalOpen] = useState(false);
	return (
		<div className="min-h-screen bg-[url('/images/concrete-texture.jpg')] bg-cover">
			<div className="container mx-auto px-4 pb-20">
				{/* Home Content */}
				<main className="mt-6">
					<div className="relative mb-8 rounded-xl overflow-hidden">
						<div className="absolute inset-0 bg-gradient-to-b from-transparent to-black/70"></div>
						<Image
							src="/mockbg.jpg"
							alt="Home Banner"
							width={1200}
							height={400}
							className="w-full h-[200px] md:h-[300px] object-cover"
						/>
						<div className="absolute bottom-0 left-0 p-4 md:p-8">
							<h2 className="text-3xl md:text-5xl font-bold text-white font-graffiti mb-2">
								Welcome to My Home
							</h2>
							<p className="text-white/80 max-w-md">
								Express yourself through digital graffiti and connect with
								friends
							</p>
							<div className="flex gap-3 mt-4">
								<Button
									className="bg-primary hover:bg-primary/90 text-white dark:text-black"
									onClick={() => setCreateWallModalOpen(true)}
								>
									<Plus className="mr-2 h-4 w-4" /> Create Wall
								</Button>
								<Button
									variant="outline"
									className="bg-black/50 text-white border-white/20 hover:bg-black/70"
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

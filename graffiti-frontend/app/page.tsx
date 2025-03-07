"use client";

import {useState} from "react";
import Image from "next/image";
import {Menu, Bell, Search, Plus} from "lucide-react";

import {Button} from "@/components/ui/button";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {MainNav} from "@/components/main-nav";
import {WallGrid} from "@/components/wall-grid";
import {MobileNav} from "@/components/mobile-nav";
import {CreateWallModal} from "@/components/create-wall-modal";
import {ThemeToggle} from "@/components/theme-toggle";

export default function HomePage() {
	const [createWallModalOpen, setCreateWallModalOpen] = useState(false);
	return (
		<div className="min-h-screen bg-[url('/images/concrete-texture.jpg')] bg-cover">
			<div className="container mx-auto px-4 pb-20">
				{/* Mobile Header */}
				<header className="flex items-center justify-between py-4 md:hidden">
					<Button variant="ghost" size="icon" className="relative">
						<Menu className="h-6 w-6" />
					</Button>
					<h1 className="text-3xl font-bold text-primary font-graffiti">
						Grafitti
					</h1>
					<div className="flex items-center gap-2">
						<ThemeToggle />
						<Button variant="ghost" size="icon" className="relative">
							<Bell className="h-6 w-6" />
							<span className="absolute top-1 right-1 h-2 w-2 rounded-full bg-destructive"></span>
						</Button>
					</div>
				</header>

				{/* Desktop Header - Hidden on mobile */}
				<header className="hidden md:flex items-center justify-between py-6">
					<div className="flex items-center gap-6">
						<h1 className="text-4xl font-bold text-primary font-graffiti">
							Grafitti
						</h1>
						<MainNav />
					</div>
					<div className="flex items-center gap-4">
						<div className="relative w-64">
							<Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
							<input
								placeholder="Search walls, friends..."
								className="w-full rounded-md border border-input bg-background px-9 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
							/>
						</div>
						<Button variant="ghost" size="icon" className="relative">
							<Bell className="h-6 w-6" />
							<span className="absolute top-1 right-1 h-2 w-2 rounded-full bg-destructive"></span>
						</Button>
						<Avatar>
							<AvatarImage src="/images/avatar.png" alt="User" />
							<AvatarFallback>LA</AvatarFallback>
						</Avatar>
					</div>
				</header>

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

			{/* Mobile Navigation */}
			<MobileNav />
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

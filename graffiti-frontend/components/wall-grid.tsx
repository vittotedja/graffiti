"use client";

import Link from "next/link";
import Image from "next/image";
import {MoreVertical, Lock, Globe, Plus, Pin} from "lucide-react";
import {useState} from "react";

import {Button} from "@/components/ui/button";
import {Card, CardContent} from "@/components/ui/card";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {CreateWallModal} from "@/components/create-wall-modal";

export function WallGrid() {
	const [createWallModalOpen, setCreateWallModalOpen] = useState(false);
	// Mock data for walls
	const walls = [
		{
			id: 1,
			title: "Birthday Wall",
			year: "2023",
			isPrivate: false,
			posts: 12,
			isPinned: true,
		},
		{
			id: 2,
			title: "Graduation",
			year: "2022",
			isPrivate: true,
			posts: 8,
			isPinned: true,
		},
		{
			id: 3,
			title: "Summer Trip",
			year: "2023",
			isPrivate: false,
			posts: 24,
			isPinned: false,
		},
		{
			id: 4,
			title: "Art Projects",
			year: "2021",
			isPrivate: false,
			posts: 15,
			isPinned: false,
		},
		{
			id: 5,
			title: "Family Reunion",
			year: "2022",
			isPrivate: true,
			posts: 18,
			isPinned: false,
		},
		{
			id: 6,
			title: "Music Festival",
			year: "2023",
			isPrivate: false,
			posts: 9,
			isPinned: false,
		},
	];

	return (
		<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{walls.map((wall) => (
				<Link href={`/wall/${wall.id}`} key={wall.id}>
					<Card className="overflow-hidden border-2 border-primary/20 bg-background/80 backdrop-blur-sm hover:shadow-lg transition-all">
						{wall.isPinned && (
							<div className="absolute top-2 left-2 z-40 bg-black/35 p-1 rounded-full">
								<Pin className="h-4 w-4 text-red-600 fill-red-600" />
							</div>
						)}
						<CardContent className="p-0">
							<div className="relative">
								<Image
									src={`/mockbday.webp`}
									alt={wall.title}
									width={400}
									height={200}
									className="w-full h-[160px] object-cover"
								/>
								<div className="absolute top-2 right-2">
									<DropdownMenu>
										<DropdownMenuTrigger asChild>
											<Button
												variant="ghost"
												size="icon"
												className="h-8 w-8 bg-black/30 text-white hover:bg-black/50 rounded-full cursor-pointer"
											>
												<MoreVertical className="h-4 w-4" />
											</Button>
										</DropdownMenuTrigger>
										<DropdownMenuContent align="end">
											<DropdownMenuItem className="cursor-pointer">
												Edit Wall
											</DropdownMenuItem>
											<DropdownMenuItem className="cursor-pointer">
												Change Privacy
											</DropdownMenuItem>
											<DropdownMenuSeparator />
											<DropdownMenuItem className="text-destructive cursor-pointer">
												Delete Wall
											</DropdownMenuItem>
										</DropdownMenuContent>
									</DropdownMenu>
								</div>
								<div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/80 to-transparent p-4">
									<h3 className="font-bold text-xl text-white font-graffiti">
										{wall.title}
									</h3>
									<div className="flex justify-between items-center mt-1">
										<span className="text-white/80 text-sm">{wall.year}</span>
										<div className="flex items-center gap-2">
											<span className="text-white/80 text-sm">
												{wall.posts} posts
											</span>
											{wall.isPrivate ? (
												<Lock className="h-4 w-4 text-white/80" />
											) : (
												<Globe className="h-4 w-4 text-white/80" />
											)}
										</div>
									</div>
								</div>
							</div>
						</CardContent>
					</Card>
				</Link>
			))}

			{/* Add New Wall Card */}
			<Card
				className="overflow-hidden border-2 border-dashed border-primary/40 bg-background/80 backdrop-blur-sm hover:bg-accent/10 transition-all h-[220px] flex items-center justify-center cursor-pointer"
				onClick={() => setCreateWallModalOpen(true)}
			>
				<CardContent className="flex flex-col items-center justify-center p-6 text-center">
					<div className="h-12 w-12 rounded-full bg-primary/10 flex items-center justify-center mb-3">
						<Plus className="h-6 w-6 text-primary" />
					</div>
					<h3 className="font-bold text-xl font-graffiti">Create New Wall</h3>
					<p className="text-muted-foreground text-sm mt-1">
						Add a new space for your memories
					</p>
				</CardContent>
			</Card>

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

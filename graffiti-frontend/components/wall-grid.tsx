"use client";

import Link from "next/link";
import Image from "next/image";
import {MoreVertical, Lock, Globe, Plus} from "lucide-react";
import {useEffect, useState} from "react";

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
import {fetchWithAuth} from "@/lib/auth";
import {formatDate} from "@/lib/date-utils";
import {Wall} from "@/types/wall";

export function WallGrid() {
	const [createWallModalOpen, setCreateWallModalOpen] = useState(false);
	const [walls, setWalls] = useState<Wall[]>([]);

	const fetchWallData = async () => {
		try {
			const response = await fetchWithAuth(
				"http://localhost:8080/api/v1/walls"
			);
			if (!response) return; // already redirected if 401

			const data = await response.json();
			console.log("Fetched wall data:", data);
			setWalls(data);
		} catch (err) {
			console.error("Failed to fetch wall data:", err);
		}
	};

	useEffect(() => {
		fetchWallData();
	}, []);

	return (
		<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{walls.length === 0 && (
				<div className="col-span-3 flex flex-col justify-center gap-2">
					<p className="text-muted-foreground">No walls found</p>
					<p className="text-muted-foreground">Start creating your walls now</p>
				</div>
			)}
			{walls.map((wall) => (
				<Link href={`/wall/${wall.id}`} key={wall.id}>
					<Card className="overflow-hidden border-2 border-primary/20 bg-background/80 h-[220px] backdrop-blur-sm hover:shadow-lg transition-all">
						{/* {wall.isPinned && (
							<div className="absolute top-2 left-2 z-40 bg-black/35 p-1 rounded-full">
								<Pin className="h-4 w-4 text-red-600 fill-red-600" />
							</div>
						)} */}
						<CardContent className="p-0">
							<div className="relative">
								<Image
									src={`/mockbday.webp`}
									alt={wall.title}
									width={400}
									height={200}
									className="w-full h-[220px] object-cover"
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
										<span className="text-white/80 text-sm">
											{formatDate(wall.created_at)}
										</span>
										<div className="flex items-center gap-2">
											{wall.is_public ? (
												<Globe className="h-4 w-4 text-white/80" />
											) : (
												<Lock className="h-4 w-4 text-white/80" />
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
			/>
		</div>
	);
}

"use client";
import {Button} from "@/components/ui/button";
import {useUser} from "@/hooks/useUser";
import {fetchWithAuth} from "@/lib/auth";
import {
	ChevronLeft,
	Globe,
	MoreVertical,
	Trash2,
	Lock,
	ArchiveRestore,
} from "lucide-react";
import {useRouter} from "next/navigation";
import React, {useEffect, useState} from "react";
import {Wall} from "@/types/wall";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {Card, CardContent} from "@/components/ui/card";
import Image from "next/image";
import {formatDate} from "@/lib/date-utils";
import Link from "next/link";
import {toast} from "sonner";

function ArchivePage() {
	const router = useRouter();
	const {user} = useUser();
	const [walls, setWalls] = useState<Wall[]>([]);

	const getArchivedWalls = async () => {
		try {
			const response = await fetchWithAuth("/api/v1/walls/archived", {
				method: "GET",
				headers: {
					"Content-Type": "application/json",
				},
			});
			if (!response.ok) {
				throw new Error("Failed to fetch archived walls");
			}
			const data = await response.json();
			setWalls(data);
		} catch (error) {
			console.error("Error fetching archived walls:", error);
		}
	};

	const unArchiveWall = async (wallId: string) => {
		try {
			const response = await fetchWithAuth(
				`/api/v1/walls/${wallId}/unarchive`,
				{
					method: "PUT",
				}
			);
			if (!response.ok) return;
			setWalls((prevWalls) => prevWalls.filter((wall) => wall.id !== wallId));
			toast.success("Wall is restored successfully");
		} catch (err) {
			console.error("Failed to restore wall:", err);
			toast.error("Failed to restore wall");
		}
	};
	useEffect(() => {
		if (!user) return;
		getArchivedWalls();
	}, [user]);

	if (!user) return;

	return (
		<div className="min-h-screen bg-background">
			<div className="container mx-auto px-4 py-8">
				<div className="mb-8">
					<div className="flex gap-2">
						<Button
							variant="ghost"
							size="icon"
							className="rounded-full w-12 h-12"
							onClick={() => router.back()}
						>
							<ChevronLeft />
						</Button>
						<h1 className="text-5xl font-bold font-graffiti">
							{user?.username}
							{`'`}s Archive
						</h1>
					</div>
					<p className="text-muted-foreground mt-2">3 wall(s)</p>
				</div>
				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
					{walls.length === 0 && (
						<div className="col-span-3 flex flex-col justify-center items-center gap-2">
							<p className="text-muted-foreground">No archived walls found</p>
						</div>
					)}
					{walls &&
						walls.length > 0 &&
						walls
							.sort((a, b) => (b.is_pinned ? 1 : 0) - (a.is_pinned ? 1 : 0))
							.map((wall) => (
								<Link key={wall.id} href={`/wall/${wall.id}`}>
									<Card className="overflow-hidden border-2 border-primary/20 bg-background/80 h-[220px] backdrop-blur-sm hover:shadow-lg transition-all">
										<CardContent className="p-0">
											<div className="relative">
												<Image
													src={wall.background_image || `/mockbday.webp`}
													alt={wall.title}
													width={400}
													height={200}
													className="w-full h-[220px] object-cover"
												/>
												{wall.user_id == user.id && (
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
																<DropdownMenuItem
																	className="cursor-pointer"
																	onClick={(e) => {
																		e.preventDefault();
																		unArchiveWall(wall.id);
																	}}
																>
																	<ArchiveRestore />
																	Restore Wall
																</DropdownMenuItem>
																<DropdownMenuSeparator />
																<DropdownMenuItem
																	className="text-destructive cursor-pointer"
																	onClick={(e) => {
																		e.preventDefault();
																	}}
																>
																	<Trash2 className="text-destructive" /> Delete
																	Wall
																</DropdownMenuItem>
															</DropdownMenuContent>
														</DropdownMenu>
													</div>
												)}
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
				</div>
			</div>
		</div>
	);
}

export default ArchivePage;

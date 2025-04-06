"use client";
import {useEffect, useState} from "react";
import {Plus, Filter, ArrowUpDown, Lock, Globe} from "lucide-react";

import {Button} from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {useUser} from "@/hooks/useUser";
import {PostGrid} from "@/components/post-grid";
import {EnhancedPostModal} from "@/components/enhanced-post-modal";
import {useParams} from "next/navigation";
import {Switch} from "@/components/ui/switch";
import {fetchWithAuth} from "@/lib/auth";
import {Wall} from "@/types/wall";
import {Post} from "@/types/post";
import {toast} from "sonner";

type SortOption = "latest" | "oldest" | "popular";
type FilterOption = "all" | "photos" | "embed";

export default function WallPage() {
	const params = useParams();
	const {user} = useUser();
	const [id, setId] = useState<string | null>(null);
	const [wallData, setWallData] = useState<Wall>();

	const fetchWallData = async (id: string) => {
		try {
			const response = await fetchWithAuth(
				"http://localhost:8080/api/v1/walls/" + id,
				{
					method: "GET",
				}
			);
			if (!response.ok) {
				throw new Error("Failed to fetch wall data");
			}
			const data = await response.json();
			// console.log(data);
			setWallData(data);
		} catch (error) {
			console.error("Error fetching wall data:", error);
		}
	};

	const fetchPostData = async (id: string) => {
		try {
			const response = await fetchWithAuth(
				"http://localhost:8080/api/v2/walls/" + id + "/posts",
				{
					method: "GET",
				}
			);
			if (!response.ok) {
				throw new Error("Failed to fetch post data");
			}
			const data = await response.json();
			setPosts(data);
		} catch (error) {
			console.error("Error fetching wall data:", error);
		}
	};

	const changePrivacy = async () => {
		if (!id) return;

		const endpoint = wallData?.is_public
			? "http://localhost:8080/api/v1/walls/" + id + "/privatize"
			: "http://localhost:8080/api/v1/walls/" + id + "/publicize";
		try {
			const response = await fetchWithAuth(endpoint, {
				method: "PUT",
			});
			if (!response.ok) {
				throw new Error("Failed to update wall privacy");
			}
			toast.success("Wall privacy updated successfully!");
			// const data = await response.json();
			// setWallData(data);
		} catch (error) {
			console.error("Error updating wall privacy:", error);
		}
	};

	useEffect(() => {
		if (params.id) {
			setId(params.id as string);
		}
		fetchWallData(params.id as string);
		fetchPostData(params.id as string);
	}, [params]);

	const [createPostModalOpen, setCreatePostModalOpen] = useState(false);
	const [sortBy, setSortBy] = useState<SortOption>("latest");
	const [filterBy, setFilterBy] = useState<FilterOption>("all");
	const [posts, setPosts] = useState<Post[]>([]);

	// Handle new post creation
	const handlePostCreated = () => {
		if (id) fetchPostData(id);
	};

	// Sort posts based on selected option
	const sortedPosts = [...posts].sort((a, b) => {
		switch (sortBy) {
			case "oldest":
				return (
					new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
				);
			case "popular":
				return b.likes_count - a.likes_count;
			case "latest":
			default:
				return (
					new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
				);
		}
	});

	// Filter posts based on selected option
	const filteredPosts = sortedPosts.filter((post) => {
		switch (filterBy) {
			case "photos":
				return post.media_url && !post.media_url.includes("video");
			case "all":
			default:
				return true;
		}
	});

	return (
		<div className="min-h-screen bg-background">
			<div className="container mx-auto px-4 py-8">
				{/* Wall Title */}
				<div className="mb-8">
					<h1 className="text-5xl font-bold font-graffiti">
						{wallData?.title}
					</h1>
					<p className="text-muted-foreground mt-2">{posts.length} post(s)</p>
				</div>

				{/* Sort and Filter Options */}
				<div className="flex gap-4 mb-8">
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button variant="outline" className="gap-2">
								<ArrowUpDown className="h-4 w-4" />
								Sort by
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent>
							<DropdownMenuItem onClick={() => setSortBy("latest")}>
								Latest
							</DropdownMenuItem>
							<DropdownMenuItem onClick={() => setSortBy("oldest")}>
								Oldest
							</DropdownMenuItem>
							<DropdownMenuItem onClick={() => setSortBy("popular")}>
								Most Popular
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>

					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button variant="outline" className="gap-2">
								<Filter className="h-4 w-4" />
								Filter
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent>
							<DropdownMenuItem onClick={() => setFilterBy("all")}>
								All Posts
							</DropdownMenuItem>
							<DropdownMenuSeparator />
							<DropdownMenuItem onClick={() => setFilterBy("photos")}>
								Photos Only
							</DropdownMenuItem>
							<DropdownMenuItem onClick={() => setFilterBy("embed")}>
								Embed Link Only
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>
					{user?.id === wallData?.user_id && (
						<div className="flex items-center gap-2">
							<Switch
								className="cursor-pointer"
								checked={!wallData?.is_public}
								onCheckedChange={() => {
									changePrivacy();
									setWallData((prev) => {
										if (prev) {
											return {...prev, is_public: !prev.is_public};
										}
										return prev;
									});
								}}
							/>
							{wallData?.is_public ? (
								<div className="flex gap-2 items-center">
									<Globe className="h-4 w-4 text-primary" />
									Public
								</div>
							) : (
								<div className="flex gap-2 items-center">
									<Lock className="h-4 w-4 text-primary" />
									Private
								</div>
							)}
						</div>
					)}
				</div>

				{/* Posts Grid */}
				<PostGrid posts={filteredPosts} />

				{/* Floating Add Post Button */}
				<Button
					className="fixed bottom-6 right-6 h-14 w-14 rounded-full shadow-lg cursor-pointer"
					onClick={() => setCreatePostModalOpen(true)}
					variant={"special"}
				>
					<Plus className="h-6 w-6" />
				</Button>

				{/* Create Post Modal */}
				{id && (
					<EnhancedPostModal
						isOpen={createPostModalOpen}
						onClose={() => setCreatePostModalOpen(false)}
						wallId={id}
						onPostCreated={handlePostCreated}
					/>
				)}
			</div>
		</div>
	);
}

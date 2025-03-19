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
import {PostGrid} from "@/components/post-grid";
import {EnhancedPostModal} from "@/components/enhanced-post-modal";
import {useParams} from "next/navigation";
import {Switch} from "@/components/ui/switch";

type SortOption = "latest" | "oldest" | "popular";
type FilterOption = "all" | "photos" | "videos" | "text";

export default function WallPage() {
	const params = useParams(); // ✅ Already synchronous
	const [id, setId] = useState<string | null>(null);

	useEffect(() => {
		if (params.id) {
			setId(params.id as string); // ✅ No async function needed
		}
	}, [params]);
	const [createPostModalOpen, setCreatePostModalOpen] = useState(false);
	const [isPrivate, setIsPrivate] = useState(false);
	const [sortBy, setSortBy] = useState<SortOption>("latest");
	const [filterBy, setFilterBy] = useState<FilterOption>("all");
	const [posts, setPosts] = useState([
		{
			id: 1,
			imageUrl: "/mockimg.jpg",
			createdAt: new Date(2024, 2, 7),
			likes: 24,
			comments: 5,
			author: {
				name: "John Doe",
				username: "johndoe",
				avatar: "/placeholder.svg?height=40&width=40",
			},
		},
		{
			id: 2,
			imageUrl: "/placeholder.svg?height=400&width=400",
			createdAt: new Date(2024, 2, 6),
			likes: 18,
			comments: 3,
			author: {
				name: "Jane Smith",
				username: "janesmith",
				avatar: "/placeholder.svg?height=40&width=40",
			},
		},
		{
			id: 3,
			imageUrl: "/placeholder.svg?height=400&width=400",
			createdAt: new Date(2024, 2, 5),
			likes: 32,
			comments: 7,
			author: {
				name: "Mike Johnson",
				username: "mikej",
				avatar: "/placeholder.svg?height=40&width=40",
			},
		},
		// Add more mock posts...
		{
			id: 4,
			imageUrl: "/placeholder.svg?height=400&width=400",
			createdAt: new Date(2024, 2, 4),
			likes: 45,
			comments: 12,
			author: {
				name: "Sarah Wilson",
				username: "sarahw",
				avatar: "/placeholder.svg?height=40&width=40",
			},
		},
		{
			id: 5,
			imageUrl: "/placeholder.svg?height=400&width=400",
			createdAt: new Date(2024, 2, 3),
			likes: 29,
			comments: 8,
			author: {
				name: "Alex Brown",
				username: "alexb",
				avatar: "/placeholder.svg?height=40&width=40",
			},
		},
		{
			id: 6,
			imageUrl: "/placeholder.svg?height=400&width=400",
			createdAt: new Date(2024, 2, 2),
			likes: 37,
			comments: 15,
			author: {
				name: "Chris Lee",
				username: "chrisl",
				avatar: "/placeholder.svg?height=40&width=40",
			},
		},
	]);

	// Handle new post creation
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	const handlePostCreated = (newPost: any) => {
		setPosts([newPost, ...posts]);
	};

	// Sort posts based on selected option
	const sortedPosts = [...posts].sort((a, b) => {
		switch (sortBy) {
			case "oldest":
				return (
					new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
				);
			case "popular":
				return b.likes - a.likes;
			case "latest":
			default:
				return (
					new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
				);
		}
	});

	// Filter posts based on selected option
	const filteredPosts = sortedPosts.filter((post) => {
		switch (filterBy) {
			case "photos":
				return post.imageUrl && !post.imageUrl.includes("video");
			case "videos":
				return post.imageUrl && post.imageUrl.includes("video");
			case "text":
				return !post.imageUrl;
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
					<h1 className="text-5xl font-bold font-graffiti">Summer Wall 2024</h1>
					<p className="text-muted-foreground mt-2">{posts.length} posts</p>
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
							<DropdownMenuItem onClick={() => setFilterBy("videos")}>
								Videos Only
							</DropdownMenuItem>
							<DropdownMenuItem onClick={() => setFilterBy("text")}>
								Text Only
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>
					<div className="flex items-center gap-2">
						<Switch
							checked={isPrivate}
							onCheckedChange={() => setIsPrivate(!isPrivate)}
						/>
						{/* {isPrivate ? "Private" : "Public"} */}
						{isPrivate ? (
							<div className="flex gap-2 items-center">
								<Lock className="h-4 w-4 text-white/80" />
								Private
							</div>
						) : (
							<div className="flex gap-2 items-center">
								<Globe className="h-4 w-4 text-white/80" />
								Public
							</div>
						)}
					</div>
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

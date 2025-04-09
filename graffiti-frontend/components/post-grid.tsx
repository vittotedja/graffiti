"use client";

import {useState, useEffect, useMemo} from "react";
import Image from "next/image";
import {formatDistanceToNow} from "date-fns";
import {Heart, MoreVertical, Trash} from "lucide-react";
import {useUser} from "@/hooks/useUser";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardFooter} from "@/components/ui/card";
import {formatFullName} from "@/lib/formatter";
import {toast} from "sonner";
import {Post} from "@/types/post";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "./ui/dropdown-menu";
import {fetchWithAuth} from "@/lib/auth";

type PostCardType = {
	post: Post;
	isWallOwner: boolean;
	onPostRemoved?: () => void;
};

interface PostGridProps {
	posts: Post[];
	isWallOwner: boolean;
	onPostRemoved?: () => void;
}

export function PostGrid({posts, isWallOwner, onPostRemoved}: PostGridProps) {
	if (!posts || posts.length === 0) {
		return (
			<div className="text-center text-muted-foreground">No posts found</div>
		);
	}
	return (
		<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{posts.map((post) => (
				<PostCard key={post.id} post={post} isWallOwner={isWallOwner} onPostRemoved={onPostRemoved} />
			))}
		</div>
	);
}

function PostCard({post, isWallOwner, onPostRemoved}: PostCardType) {
	const {user} = useUser();
	const [liked, setLiked] = useState(false);
	const [likeCount, setLikeCount] = useState(post.likes_count);
	const [tiktokHtml, setTiktokHtml] = useState(null);

	const handleLike = async () => {
		try {
			const response = await fetchWithAuth(`/api/v1/likes`, {
				method: "POST",
				body: JSON.stringify({
					post_id: post.id, // <-- adjust to match your backend expected JSON field
				}),
			});

			if (!response.ok) {
				throw new Error("Failed to toggle like");
			}

			const data = await response.json();

			if (data.message == "Post liked successfully") {
				setLiked(true);
				setLikeCount((prev) => prev + 1);
			} else if (data.message == "Post unliked successfully") {
				setLiked(false);
				setLikeCount((prev) => Math.max(0, prev - 1));
			}
		} catch (error) {
			console.error("Error toggling like:", error);
		}
	};

	useEffect(() => {
		if (!user) return;
		const haveLikedPost = async () => {
			try {
				const response = await fetchWithAuth(`/api/v1/likes/${post.id}`, {
					method: "GET",
				});
				if (response.ok) {
					const data = await response.json();
					setLiked(data.liked);
				}
			} catch (error) {
				console.error("Error checking like status:", error);
			}
		};
		haveLikedPost();
	}, [post.id, user]);

	// Compute embed data once from the media URL
	const embedData = useMemo(
		() => getEmbedUrl(post.media_url),
		[post.media_url]
	);

	// Fetch TikTok embed only when needed
	useEffect(() => {
		if (embedData.type === "tiktok" && !tiktokHtml) {
			processTiktokPreview(post.media_url);
		}
	}, [post.media_url, embedData.type, tiktokHtml]);

	// Moved TikTok processing to a helper function
	const processTiktokPreview = async (mediaUrl: string) => {
		try {
			const response = await fetch(
				`https://www.tiktok.com/oembed?url=${mediaUrl}`
			);
			const data = await response.json();
			if (data.html) {
				setTiktokHtml(data.html);
				// Dynamically load the TikTok script after setting the HTML
				setTimeout(() => {
					const script = document.createElement("script");
					script.src = "https://www.tiktok.com/embed.js";
					script.async = true;
					document.body.appendChild(script);
				}, 500);
			} else {
				toast.error("TikTok embed failed", {
					description: "Could not fetch the TikTok embed.",
				});
			}
		} catch (error) {
			toast.error("TikTok fetch error", {
				description: "Failed to fetch the TikTok embed URL.",
			});
			console.error("Error fetching TikTok embed:", error);
		}
	};

	// Renders the correct iframe or embed based on the platform type
	const renderIframe = () => {
		switch (embedData.type) {
			case "youtube":
				return (
					<iframe
						src={embedData.embedUrl}
						className="w-full h-full aspect-video"
						allowFullScreen
						allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
					></iframe>
				);
			case "spotify":
				return (
					<iframe
						className="border-r-[12px]"
						src={embedData.embedUrl}
						width="100%"
						height="352"
						allowFullScreen
						allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture"
						loading="lazy"
					></iframe>
				);
			case "tiktok":
				if (!tiktokHtml) return <p>Loading TikTok preview...</p>;
				return (
					<div className="w-full tiktok-div">
						<div dangerouslySetInnerHTML={{__html: tiktokHtml}} />
					</div>
				);
			default:
				return (
					<div className="p-4 text-center text-sm text-muted-foreground">
						<p>Preview for {post.media_url}</p>
						<p className="mt-2">
							(Non-YouTube/Spotify previews would be implemented based on
							specific platform APIs)
						</p>
					</div>
				);
		}
	};

	const removePost = async () => {
		try {
			const response = await fetchWithAuth(`/api/v1/posts/${post.id}`, {
				method: "DELETE",
			});
			if (response.ok) {
				toast.success("Post removed successfully");
				// Call the callback function to refresh posts if it exists
				if (onPostRemoved) {
					onPostRemoved();
				}
			}
		} catch (error) {
			console.error(error);
			toast.error("Error removing post");
		}
	};

	if (!user) return;

	return (
		<div className="relative">
			{(isWallOwner || user.username === post.username) && (
				<DropdownMenu>
					<DropdownMenuTrigger asChild className="relative">
						<Button
							variant="ghost"
							size="icon"
							className="h-8 w-8 absolute top-2 right-2 z-10 bg-gray-700"
						>
							<MoreVertical className="h-4 w-4" />
						</Button>
					</DropdownMenuTrigger>
					<DropdownMenuContent align="end">
						<DropdownMenuItem
							className="text-red-500 cursor-pointer"
							onClick={removePost}
						>
							<Trash className="text-red-500" />
							Remove Post
						</DropdownMenuItem>
					</DropdownMenuContent>
				</DropdownMenu>
			)}
			<Card className="overflow-hidden border border-border/40 bg-background/60 backdrop-blur-sm hover:bg-background/80 transition-colors shadow-cyan-200">
				<CardContent className="p-0">
					{post.post_type === "embed_link" && renderIframe()}
					{post.post_type === "media" && post.media_url && (
						<div className="relative aspect-square">
							<Image
								src={post.media_url || "/placeholder.svg"}
								alt="Post image"
								fill
								sizes="100%"
								className="object-cover"
							/>
						</div>
					)}
				</CardContent>
				<CardFooter className="p-4 flex flex-col gap-3">
					<div className="flex items-center justify-between w-full">
						<div className="flex items-center gap-2">
							<Avatar className="h-8 w-8">
								<AvatarImage src={post.profile_picture} alt={post.username} />
								<AvatarFallback>{formatFullName(post.fullname)}</AvatarFallback>
							</Avatar>
							<div className="text-sm">
								<div className="font-medium">@{post.username}</div>
								<div className="text-xs text-muted-foreground">
									{formatDistanceToNow(new Date(post.created_at), {
										addSuffix: true,
									})}
								</div>
							</div>
						</div>
						<div className="flex gap-2 items-center">
							{likeCount}
							<Button
								variant="ghost"
								size="icon"
								className={liked ? "text-red-500" : ""}
								onClick={handleLike}
							>
								<Heart className={`h-5 w-5 ${liked ? "fill-red-500" : ""}`} />
							</Button>
						</div>
					</div>
				</CardFooter>
			</Card>
		</div>
	);
}

// Helper to determine embed URL and platform type
function getEmbedUrl(url: string) {
	// YouTube
	if (url.includes("youtube.com") || url.includes("youtu.be")) {
		return {embedUrl: url, type: "youtube"};
	}
	// Spotify
	if (url.includes("spotify.com")) {
		if (url.includes("/embed/")) {
			return {embedUrl: url, type: "spotify"};
		} else if (
			url.includes("/track/") ||
			url.includes("/album/") ||
			url.includes("/playlist/")
		) {
			const parts = url.split(".com/");
			if (parts.length > 1) {
				const path = parts[1];
				return {
					embedUrl: `https://open.spotify.com/embed/${path}`,
					type: "spotify",
				};
			}
		}
	}
	// TikTok
	if (url.includes("tiktok.com")) {
		return {embedUrl: url, type: "tiktok"};
	}
	// Other
	return {embedUrl: url, type: "others"};
}

export default PostCard;

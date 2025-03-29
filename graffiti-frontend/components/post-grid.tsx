"use client";

import {useState} from "react";
import Image from "next/image";
import {formatDistanceToNow} from "date-fns";
import {Heart} from "lucide-react";

import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardFooter} from "@/components/ui/card";
import {Post} from "@/types/post";
import {formatFullName} from "@/lib/formatter";

interface PostGridProps {
	posts: Post[];
}

export function PostGrid({posts}: PostGridProps) {
	if (!posts || posts.length === 0) {
		return (
			<div className="text-center text-muted-foreground">No posts found</div>
		);
	}
	return (
		<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{posts.map((post) => (
				<PostCard key={post.id} post={post} />
			))}
		</div>
	);
}

function PostCard({post}: {post: Post}) {
	const [liked, setLiked] = useState(false);
	const [likeCount, setLikeCount] = useState(post.likes_count);

	const handleLike = () => {
		if (liked) {
			setLikeCount(likeCount - 1);
		} else {
			setLikeCount(likeCount + 1);
		}
		setLiked(!liked);
	};

	return (
		<Card className="overflow-hidden border border-border/40 bg-background/60 backdrop-blur-sm hover:bg-background/80 transition-colors shadow-cyan-200">
			<CardContent className="p-0">
				{post.media_url && (
					<div className="relative aspect-square">
						<Image
							src={post.media_url || "/placeholder.svg"}
							alt="Post image"
							fill
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
					<Button
						variant="ghost"
						size="icon"
						className={`${liked ? "text-red-500" : ""}`}
						onClick={handleLike}
					>
						<Heart className={`h-5 w-5 ${liked ? "fill-red-500" : ""}`} />
					</Button>
				</div>
			</CardFooter>
		</Card>
	);
}

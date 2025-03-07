"use client";

import {useState} from "react";
import Image from "next/image";
import {formatDistanceToNow} from "date-fns";
import {Heart} from "lucide-react";

import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardFooter} from "@/components/ui/card";

interface Post {
	id: number;
	content: string;
	imageUrl: string;
	createdAt: Date;
	likes: number;
	comments: number;
	author: {
		name: string;
		username: string;
		avatar: string;
	};
}

interface PostGridProps {
	posts: Post[];
}

export function PostGrid({posts}: PostGridProps) {
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
	const [likeCount, setLikeCount] = useState(post.likes);

	const handleLike = () => {
		if (liked) {
			setLikeCount(likeCount - 1);
		} else {
			setLikeCount(likeCount + 1);
		}
		setLiked(!liked);
	};

	return (
		<Card className="overflow-hidden border border-border/40 bg-background/60 backdrop-blur-sm hover:bg-background/80 transition-colors">
			<CardContent className="p-0">
				{post.imageUrl && (
					<div className="relative aspect-square">
						<Image
							src={post.imageUrl || "/placeholder.svg"}
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
							<AvatarImage src={post.author.avatar} alt={post.author.name} />
							<AvatarFallback>{post.author.name.charAt(0)}</AvatarFallback>
						</Avatar>
						<div className="text-sm">
							<div className="font-medium">@{post.author.username}</div>
							<div className="text-xs text-muted-foreground">
								{formatDistanceToNow(new Date(post.createdAt), {
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
				{post.content && (
					<p className="text-sm text-muted-foreground">{post.content}</p>
				)}
			</CardFooter>
		</Card>
	);
}

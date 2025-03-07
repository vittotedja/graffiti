"use client";
import {useState} from "react";
import Image from "next/image";
import {formatDistanceToNow} from "date-fns";
import {Heart, MessageSquare, Share2, MoreHorizontal} from "lucide-react";

import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardFooter, CardHeader} from "@/components/ui/card";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface PostCardProps {
	post: {
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
	};
}

export function PostCard({post}: PostCardProps) {
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

	// Format the date properly, handling both string dates and Date objects
	const formattedDate =
		post.createdAt instanceof Date
			? formatDistanceToNow(post.createdAt, {addSuffix: true})
			: formatDistanceToNow(new Date(post.createdAt), {addSuffix: true});

	return (
		<Card className="overflow-hidden border-2 border-primary/20 bg-black/5 backdrop-blur-sm">
			<CardHeader className="p-4 flex flex-row items-center space-y-0 gap-2">
				<Avatar>
					<AvatarImage src={post.author.avatar} alt={post.author.name} />
					<AvatarFallback>{post.author.name.charAt(0)}</AvatarFallback>
				</Avatar>
				<div className="flex-1">
					<div className="font-medium">{post.author.name}</div>
					<div className="text-xs text-muted-foreground">
						@{post.author.username}
					</div>
				</div>
				<div className="text-xs text-muted-foreground">{formattedDate}</div>
				<DropdownMenu>
					<DropdownMenuTrigger asChild>
						<Button variant="ghost" size="icon" className="h-8 w-8">
							<MoreHorizontal className="h-4 w-4" />
							<span className="sr-only">More options</span>
						</Button>
					</DropdownMenuTrigger>
					<DropdownMenuContent align="end">
						<DropdownMenuItem>Report post</DropdownMenuItem>
						<DropdownMenuSeparator />
						<DropdownMenuItem>Copy link</DropdownMenuItem>
					</DropdownMenuContent>
				</DropdownMenu>
			</CardHeader>
			<CardContent className="p-0">
				{post.imageUrl && (
					<div className="relative w-full">
						<Image
							src={post.imageUrl || "/placeholder.svg"}
							alt="Post image"
							width={600}
							height={400}
							className="w-full object-cover max-h-[500px]"
						/>
					</div>
				)}
				{post.content && (
					<div className="p-4 pt-2">
						<p>{post.content}</p>
					</div>
				)}
			</CardContent>
			<CardFooter className="p-4 pt-2 flex justify-between">
				<div className="flex gap-4">
					<Button
						variant="ghost"
						size="sm"
						className={`flex items-center gap-1 ${liked ? "text-red-500" : ""}`}
						onClick={handleLike}
					>
						<Heart className={`h-5 w-5 ${liked ? "fill-red-500" : ""}`} />
						<span>{likeCount}</span>
					</Button>
					<Button variant="ghost" size="sm" className="flex items-center gap-1">
						<MessageSquare className="h-5 w-5" />
						<span>{post.comments}</span>
					</Button>
				</div>
				<Button variant="ghost" size="sm" className="flex items-center gap-1">
					<Share2 className="h-5 w-5" />
					<span>Share</span>
				</Button>
			</CardFooter>
		</Card>
	);
}

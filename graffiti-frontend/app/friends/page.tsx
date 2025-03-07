import Link from "next/link";
import {ChevronLeft, UserPlus, MoreVertical, Search} from "lucide-react";

import {Button} from "@/components/ui/button";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Card, CardContent, CardHeader} from "@/components/ui/card";
import {Tabs, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {Input} from "@/components/ui/input";
import {MobileNav} from "@/components/mobile-nav";

export default function FriendsPage() {
	// Mock data for friends
	const friends = [
		{
			id: 1,
			name: "Friend Name 1",
			username: "friendname1",
			avatar: "/placeholder.svg?height=40&width=40",
			status: "friended",
		},
		{
			id: 2,
			name: "Friend Name 2",
			username: "friendname2",
			avatar: "/placeholder.svg?height=40&width=40",
			status: "friended",
		},
		{
			id: 3,
			name: "Friend Name 3",
			username: "friendname3",
			avatar: "/placeholder.svg?height=40&width=40",
			status: "friended",
		},
		{
			id: 4,
			name: "Friend Name 4",
			username: "friendname4",
			avatar: "/placeholder.svg?height=40&width=40",
			status: "pending",
		},
		{
			id: 5,
			name: "Friend Name 5",
			username: "friendname5",
			avatar: "/placeholder.svg?height=40&width=40",
			status: "pending",
		},
	];

	return (
		<div className="min-h-screen bg-[url('/images/concrete-texture.jpg')] bg-cover">
			<div className="container mx-auto px-4 pb-20">
				{/* Header */}
				<header className="py-4 flex items-center gap-2">
					<Link href="/">
						<Button variant="ghost" size="icon" className="rounded-full">
							<ChevronLeft className="h-6 w-6" />
						</Button>
					</Link>
					<h1 className="text-2xl md:text-3xl font-bold font-graffiti">
						Friends
					</h1>
				</header>

				{/* Search */}
				<div className="relative mb-6">
					<Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
					<Input
						placeholder="Search friends..."
						className="pl-9 bg-black/5 backdrop-blur-sm border-2 border-primary/20"
					/>
				</div>

				{/* Friends List */}
				<Card className="border-2 border-primary/20 bg-black/5 backdrop-blur-sm">
					<CardHeader className="px-4 py-3 border-b">
						<Tabs defaultValue="all" className="w-full">
							<TabsList className="grid w-full grid-cols-3">
								<TabsTrigger value="all">All Friends</TabsTrigger>
								<TabsTrigger value="pending">Pending</TabsTrigger>
								<TabsTrigger value="suggestions">Suggestions</TabsTrigger>
							</TabsList>
						</Tabs>
					</CardHeader>
					<CardContent className="p-0">
						<div className="divide-y">
							{friends.map((friend) => (
								<div
									key={friend.id}
									className="flex items-center justify-between p-4 hover:bg-accent/50"
								>
									<div className="flex items-center gap-3">
										<Avatar>
											<AvatarImage src={friend.avatar} alt={friend.name} />
											<AvatarFallback>{friend.name.charAt(0)}</AvatarFallback>
										</Avatar>
										<div>
											<div className="font-medium">{friend.name}</div>
											<div className="text-xs text-muted-foreground">
												@{friend.username}
											</div>
										</div>
									</div>

									{friend.status === "friended" ? (
										<DropdownMenu>
											<DropdownMenuTrigger asChild>
												<Button variant="ghost" size="icon" className="h-8 w-8">
													<MoreVertical className="h-4 w-4" />
												</Button>
											</DropdownMenuTrigger>
											<DropdownMenuContent align="end">
												<DropdownMenuItem>View Profile</DropdownMenuItem>
												<DropdownMenuItem>Message</DropdownMenuItem>
												<DropdownMenuSeparator />
												<DropdownMenuItem>Block</DropdownMenuItem>
												<DropdownMenuItem className="text-destructive">
													Remove Friend
												</DropdownMenuItem>
											</DropdownMenuContent>
										</DropdownMenu>
									) : (
										<Button size="sm">
											<UserPlus className="h-4 w-4 mr-2" />
											Accept
										</Button>
									)}
								</div>
							))}
						</div>
					</CardContent>
				</Card>
			</div>

			{/* Mobile Navigation */}
			<MobileNav />
		</div>
	);
}

"use client";
import Link from "next/link";
import {ChevronLeft, Search, User, Clock, UserRoundPlus} from "lucide-react";

import {Button} from "@/components/ui/button";
import {Card, CardContent, CardHeader} from "@/components/ui/card";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {Input} from "@/components/ui/input";
import {MobileNav} from "@/components/mobile-nav";
import {useUser} from "@/hooks/useUser";
import PendingFriendsList from "./pending-friendlist";
import RequestedFriendsList from "./sent-friendlist";
import FriendsList from "./friend-list";

export default function FriendsPage() {
	const {user, loading} = useUser();

	if (loading) return <p>Loading...</p>;
	if (!user) return <p>Not logged in</p>;

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
						placeholder="Search or Add new friends..."
						className="pl-9 bg-black/5 border-2 border-primary/20"
					/>
				</div>

				{/* Friends List */}
				<Card className="border-2 border-primary/20 bg-black/5 backdrop-blur-sm">
					<Tabs defaultValue="all" className="w-full">
						<CardHeader className="px-4 py-3 border-b">
							<TabsList className="grid w-full grid-cols-3">
								<TabsTrigger value="all">
									<User />
									Friends
								</TabsTrigger>
								<TabsTrigger value="pending">
									<Clock />
									Pending
								</TabsTrigger>
								<TabsTrigger value="requested">
									<UserRoundPlus />
									Sent
								</TabsTrigger>
							</TabsList>
						</CardHeader>

						<CardContent className="p-0">
							<TabsContent value="all">
								<FriendsList />
							</TabsContent>
							<TabsContent value="pending">
								<PendingFriendsList />
							</TabsContent>
							<TabsContent value="requested">
								<RequestedFriendsList />
							</TabsContent>
						</CardContent>
					</Tabs>
				</Card>
			</div>

			{/* Mobile Navigation */}
			<MobileNav />
		</div>
	);
}

import SuggestedFriends from "@/components/suggested-friends";
import TrendingUsers from "@/components/trending-users";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";

export default function DiscoverPage() {
	return (
		<div className="min-h-screen">
			<div className="container mx-auto px-4 pb-20">
				<div className="space-y-2 pb-4">
					<h1 className="text-3xl font-bold tracking-tight">Discover</h1>
					<p className="text-muted-foreground">
						Find new friends and connections based on your network
					</p>
				</div>

				<Tabs defaultValue="suggested" className="w-full">
					<TabsList className="grid w-full grid-cols-3 mb-8">
						<TabsTrigger value="suggested">Suggested For You</TabsTrigger>
						<TabsTrigger value="trending">Trending Users</TabsTrigger>
						<TabsTrigger value="explore">Explore Walls</TabsTrigger>
					</TabsList>
					<TabsContent value="suggested" className="space-y-6">
						<SuggestedFriends />
					</TabsContent>
					<TabsContent value="trending" className="space-y-6">
						<TrendingUsers />
					</TabsContent>
					<TabsContent value="explore" className="space-y-6">
						<div className="text-center py-12">
							<h3 className="text-lg font-medium">
								Explore popular walls coming soon!
							</h3>
							<p className="text-muted-foreground mt-2">
								We&apos;re working on bringing you the most creative walls from
								our community.
							</p>
						</div>
					</TabsContent>
				</Tabs>
			</div>
		</div>
	);
}

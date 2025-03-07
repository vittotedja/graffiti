import Link from "next/link";
import {Home, Users, PlusSquare, Bell, User} from "lucide-react";

export function MobileNav() {
	return (
		<div className="fixed bottom-0 left-0 right-0 border-t bg-background/80 backdrop-blur-md md:hidden">
			<div className="flex items-center justify-around p-2">
				<Link href="/" className="flex flex-col items-center p-2">
					<Home className="h-6 w-6" />
					<span className="text-xs mt-1">Home</span>
				</Link>
				<Link href="/friends" className="flex flex-col items-center p-2">
					<Users className="h-6 w-6" />
					<span className="text-xs mt-1">Friends</span>
				</Link>
				<Link href="/create" className="flex flex-col items-center p-2">
					<div className="bg-primary text-primary-foreground p-2 rounded-full">
						<PlusSquare className="h-6 w-6" />
					</div>
				</Link>
				<Link href="/notifications" className="flex flex-col items-center p-2">
					<div className="relative">
						<Bell className="h-6 w-6" />
						<span className="absolute top-0 right-0 h-2 w-2 rounded-full bg-destructive"></span>
					</div>
					<span className="text-xs mt-1">Alerts</span>
				</Link>
				<Link href="/profile" className="flex flex-col items-center p-2">
					<User className="h-6 w-6" />
					<span className="text-xs mt-1">Profile</span>
				</Link>
			</div>
		</div>
	);
}

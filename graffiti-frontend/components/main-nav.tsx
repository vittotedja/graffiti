import Link from "next/link";
import {Home, Users, Search, Bell} from "lucide-react";
import {ThemeToggle} from "@/components/theme-toggle";

export function MainNav() {
	return (
		<nav className="hidden md:flex items-center gap-6">
			<Link
				href="/"
				className="text-lg font-medium transition-colors hover:text-primary flex items-center gap-1"
			>
				<Home className="h-5 w-5" />
				Home
			</Link>
			<Link
				href="/friends"
				className="text-lg font-medium text-muted-foreground transition-colors hover:text-primary flex items-center gap-1"
			>
				<Users className="h-5 w-5" />
				Friends
			</Link>
			<Link
				href="/search"
				className="text-lg font-medium text-muted-foreground transition-colors hover:text-primary flex items-center gap-1"
			>
				<Search className="h-5 w-5" />
				Explore
			</Link>
			<Link
				href="/notifications"
				className="text-lg font-medium text-muted-foreground transition-colors hover:text-primary flex items-center gap-1 relative"
			>
				<Bell className="h-6 w-6" />
				<span className="absolute top-1 right-1 h-2 w-2 rounded-full bg-destructive"></span>
			</Link>
			<ThemeToggle />
		</nav>
	);
}

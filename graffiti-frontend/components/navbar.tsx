"use client";

import {useState, useEffect} from "react";
import Link from "next/link";
import {usePathname} from "next/navigation";
import {Bell, Search, Menu, X, Home, Users, Compass} from "lucide-react";

import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {ThemeToggle} from "@/components/theme-toggle";
import {
	Sheet,
	SheetContent,
	SheetHeader,
	SheetTitle,
	SheetTrigger,
	SheetClose,
} from "@/components/ui/sheet";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {formatFullName} from "@/lib/formatter";
import {useUser} from "@/hooks/useUser";

export function Navbar() {
	const pathname = usePathname();
	const {user, loading} = useUser(true);
	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	const [hasNotifications, setHasNotifications] = useState(true);
	const [isScrolled, setIsScrolled] = useState(false);
	const [isMobileSearchOpen, setIsMobileSearchOpen] = useState(false);

	// Check if the current path matches the link
	const isActive = (path: string) => {
		return pathname === path;
	};

	// Handle scroll effect for navbar
	useEffect(() => {
		const handleScroll = () => {
			setIsScrolled(window.scrollY > 10);
		};

		window.addEventListener("scroll", handleScroll);
		return () => window.removeEventListener("scroll", handleScroll);
	}, []);

	if (pathname.includes("/login")) return null;
	if (loading) return null;
	if (!user) return null;
	return (
		<header
			className={`sticky top-0 z-50 w-full transition-all duration-200 ${
				isScrolled
					? "bg-background/80 backdrop-blur-md shadow-sm"
					: "bg-background"
			}`}
		>
			<div className="container mx-auto px-4">
				<div className="flex h-16 items-center justify-between">
					{/* Logo and Mobile Menu */}
					<div className="flex items-center">
						{/* Mobile Menu */}
						<Sheet>
							<SheetTrigger asChild className="mr-2 md:hidden">
								<Button variant="ghost" size="icon" aria-label="Menu">
									<Menu className="h-5 w-5" />
								</Button>
							</SheetTrigger>
							<SheetContent side="left" className="w-[250px] sm:w-[300px]">
								<SheetHeader className="mb-6">
									<SheetTitle className="text-2xl font-bold font-graffiti text-primary">
										Graffiti
									</SheetTitle>
								</SheetHeader>
								<nav className="flex flex-col gap-4">
									<SheetClose asChild>
										<Link
											href="/"
											className={`flex items-center gap-2 px-2 py-2 rounded-md text-lg ${
												isActive("/")
													? "bg-accent font-medium"
													: "hover:bg-accent/50"
											}`}
										>
											<Home className="h-5 w-5" />
											Home
										</Link>
									</SheetClose>
									<SheetClose asChild>
										<Link
											href="/friends"
											className={`flex items-center gap-2 px-2 py-2 rounded-md text-lg ${
												isActive("/friends")
													? "bg-accent font-medium"
													: "hover:bg-accent/50"
											}`}
										>
											<Users className="h-5 w-5" />
											Friends
										</Link>
									</SheetClose>
									<SheetClose asChild>
										<Link
											href="/discover"
											className={`flex items-center gap-2 px-2 py-2 rounded-md text-lg ${
												isActive("/discover")
													? "bg-accent font-medium"
													: "hover:bg-accent/50"
											}`}
										>
											<Compass className="h-5 w-5" />
											Discover
										</Link>
									</SheetClose>
									<div className="mt-auto pt-6">
										<ThemeToggle />
									</div>
								</nav>
							</SheetContent>
						</Sheet>

						{/* Logo */}
						<Link href="/" className="flex items-center">
							<span className="text-2xl font-bold text-primary font-graffiti">
								Graffiti
							</span>
						</Link>

						{/* Desktop Navigation */}
						<nav className="hidden md:flex items-center ml-10 space-x-6">
							<Link
								href="/friends"
								className={`text-base transition-colors hover:text-primary ${
									isActive("/friends")
										? "font-medium text-foreground"
										: "text-muted-foreground"
								}`}
							>
								Friends
							</Link>
							<Link
								href="/discover"
								className={`text-base transition-colors hover:text-primary ${
									isActive("/discover")
										? "font-medium text-foreground"
										: "text-muted-foreground"
								}`}
							>
								Discover
							</Link>
						</nav>
					</div>

					{/* Search, Notifications, and Profile */}
					<div className="flex items-center gap-2">
						{/* Desktop Search */}
						<div className="hidden md:flex relative w-64">
							<Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
							<Input
								placeholder="Search walls, friends..."
								className="pl-8 bg-background"
							/>
						</div>

						{/* Mobile Search Toggle */}
						<Button
							variant="ghost"
							size="icon"
							className="md:hidden"
							onClick={() => setIsMobileSearchOpen(!isMobileSearchOpen)}
							aria-label="Search"
						>
							{isMobileSearchOpen ? (
								<X className="h-5 w-5" />
							) : (
								<Search className="h-5 w-5" />
							)}
						</Button>

						{/* Notifications */}
						<Link href="/notifications">
							<Button
								variant="ghost"
								size="icon"
								className="relative"
								aria-label="Notifications"
							>
								<Bell className="h-5 w-5" />
								{hasNotifications && (
									<span className="absolute top-1 right-1 h-2 w-2 rounded-full bg-destructive"></span>
								)}
							</Button>
						</Link>

						{/* Theme Toggle (Desktop) */}
						<div className="hidden md:block">
							<ThemeToggle />
						</div>

						{/* Profile */}
						<DropdownMenu>
							<DropdownMenuTrigger asChild>
								<Button
									variant="ghost"
									size="icon"
									className="rounded-full h-8 w-8 ml-1"
									aria-label="Profile"
								>
									<Avatar className="h-8 w-8">
										<AvatarImage
											src={
												user.profile_picture ||
												"/placeholder.svg?height=32&width=32"
											}
											alt="User"
										/>
										<AvatarFallback>
											{formatFullName(user.fullname)}
											{/* VT */}
										</AvatarFallback>
									</Avatar>
								</Button>
							</DropdownMenuTrigger>
							<DropdownMenuContent align="end">
								<DropdownMenuItem asChild>
									<Link href="/">Profile</Link>
								</DropdownMenuItem>
								{/* <DropdownMenuItem asChild>
									<Link href="/settings">Settings</Link>
								</DropdownMenuItem> */}
								<DropdownMenuSeparator />
								<DropdownMenuItem>Sign out</DropdownMenuItem>
							</DropdownMenuContent>
						</DropdownMenu>
					</div>
				</div>

				{/* Mobile Search Bar (Expandable) */}
				{isMobileSearchOpen && (
					<div className="pb-3 md:hidden">
						<div className="relative">
							<Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
							<Input
								placeholder="Search walls, friends..."
								className="pl-8 bg-background w-full"
								autoFocus
							/>
						</div>
					</div>
				)}
			</div>
		</header>
	);
}

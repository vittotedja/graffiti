import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Button} from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {Input} from "@/components/ui/input";
import {fetchWithAuth} from "@/lib/auth";
import {formatFullName} from "@/lib/formatter";
import {User} from "@/types/user";
import {MoreVertical, Search} from "lucide-react";
import Link from "next/link";
import {useEffect, useState} from "react";

export default function SearchFriends() {
	const [searchTerm, setSearchTerm] = useState<string>("");
	const [debouncedQuery, setDebouncedQuery] = useState("");
	const [userList, setUserList] = useState<User[]>([]);

	// Debounce: wait 500ms after typing stops
	useEffect(() => {
		const handler = setTimeout(() => {
			setDebouncedQuery(searchTerm);
		}, 500); // delay in ms

		return () => {
			clearTimeout(handler); // clear timeout if query changes
		};
	}, [searchTerm]);

	const searchFunction = async (debouncedQuery: string) => {
		if (debouncedQuery.length <= 0) return;
		const response = await fetchWithAuth(
			"http://localhost:8080/api/v1/users/search",
			{
				method: "POST",
				body: JSON.stringify({
					search_term: debouncedQuery,
				}),
			}
		);
		if (!response.ok) return;
		const data = await response.json();
		setUserList(data);
	};

	useEffect(() => {
		searchFunction(debouncedQuery);
	}, [debouncedQuery]);

	return (
		<div className="relative mb-6">
			<Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
			<Input
				placeholder="Search or Add new friends..."
				onChange={(e) => setSearchTerm(e.target.value)}
				className="pl-9 bg-black/5 border-2 border-primary/20"
			/>
			{searchTerm && (
				<div className="w-full rounded-sm shadow-md flex items-center justify-between hover:bg-accent/50 divide-y">
					{userList.length > 0 &&
						userList.map((user: User) => (
							<div
								key={user.id}
								className="w-full flex items-center justify-between p-4 hover:bg-accent/50"
							>
								<Link
									href={`/profile/${user.id}`}
									className="flex items-center gap-3 hover:underline cursor-pointer"
								>
									<Avatar>
										<AvatarImage
											src={user.profile_picture}
											alt={user.fullname}
										/>
										<AvatarFallback>
											{formatFullName(user.fullname)}
										</AvatarFallback>
									</Avatar>
									<div>
										<div className="font-medium">{user.fullname}</div>
										<div className="text-xs text-muted-foreground">
											@{user.username}
										</div>
									</div>
								</Link>
								<DropdownMenu>
									<DropdownMenuTrigger asChild>
										<Button variant="ghost" size="icon" className="h-8 w-8">
											<MoreVertical className="h-4 w-4" />
										</Button>
									</DropdownMenuTrigger>
									<DropdownMenuContent align="end">
										<DropdownMenuItem>
											<Link href={`/profile/${user.id}`}>View Profile</Link>
										</DropdownMenuItem>
										<DropdownMenuSeparator />
										<DropdownMenuItem>Add Friend</DropdownMenuItem>
										<DropdownMenuItem className="text-destructive">
											Remove Friend
										</DropdownMenuItem>
										<DropdownMenuItem>Block</DropdownMenuItem>
									</DropdownMenuContent>
								</DropdownMenu>
							</div>
						))}
				</div>
			)}
		</div>
	);
}

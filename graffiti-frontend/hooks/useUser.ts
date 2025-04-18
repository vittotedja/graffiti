// hooks/useUser.ts
import {fetchWithAuth} from "@/lib/auth";
import {User} from "@/types/user";
import {useEffect, useState} from "react";
import {usePathname, useRouter} from "next/navigation";

export function useUser(redirectIfNull = false) {
	const [user, setUser] = useState<User | null>(null);
	const [loading, setLoading] = useState(true);
	const router = useRouter();
	const pathname = usePathname();

	useEffect(() => {
		if (pathname === "/login") {
			setLoading(false);
			return;
		}

		fetchWithAuth("/api/v1/auth/me", {
			method: "POST",
		})
			.then((res) => {
				if (!res.ok) throw new Error("Unauthorized");
				return res.json();
			})
			.then((data) => {
				setUser(data.user);
			})
			.catch(() => {
				setUser(null);
				if (redirectIfNull && pathname !== "/login") {
					router.push("/login");
				}
			})
			.finally(() => setLoading(false));
	}, [redirectIfNull, router, pathname]);

	return {user, loading};
}

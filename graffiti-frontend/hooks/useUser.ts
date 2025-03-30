// hooks/useUser.ts
import {fetchWithAuth} from "@/lib/auth";
import {User} from "@/types/user";
import {useEffect, useState} from "react";

export function useUser() {
	const [user, setUser] = useState<User | null>(null);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		fetchWithAuth("http://localhost:8080/api/v1/auth/me", {
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
			})
			.finally(() => setLoading(false));
	}, []);

	return {user, loading};
}

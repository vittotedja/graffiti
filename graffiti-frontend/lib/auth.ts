export async function fetchWithAuth(url: string, options: RequestInit = {}) {
	const res = await fetch(url, {
		...options,
		credentials: "include", // include cookies (JWT)
	});

	if (res.status === 401) {
		if (typeof window !== "undefined") {
			window.location.href = "/login";
		}
		throw new Error("Unauthorized");
	}

	return res;
}

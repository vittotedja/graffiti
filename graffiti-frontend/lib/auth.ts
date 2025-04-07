const apiUrl = process.env.NEXT_PUBLIC_API_URL;

export async function fetchWithAuth(url: string, options: RequestInit = {}) {
	const res = await fetch(apiUrl + url, {
		...options,
		credentials: "include", // include cookies (JWT)
	});

	if (res.status === 401) {
		console.log("hello");
		if (typeof window !== "undefined") {
			window.location.href = "/login";
		}
		throw new Error("Unauthorized");
	}

	return res;
}

export function formatDate(
	date: string | Date,
	locale = "en-US",
	options?: Intl.DateTimeFormatOptions
): string {
	return new Intl.DateTimeFormat(
		locale,
		options || {
			year: "numeric",
			month: "short",
			day: "numeric",
		}
	).format(new Date(date));
}

export function formatFullName(fullname: string | undefined | null): string {
	// Safety check - return placeholder if fullname is missing
	if (!fullname) return 'NA';
	
	try {
		const separatedNames = fullname.split(" ");
		const firstName = separatedNames[0];
		const lastName = separatedNames[separatedNames.length - 1];
		
		if (separatedNames.length >= 2) {
			return `${firstName.charAt(0).toUpperCase()}${lastName
				.charAt(0)
				.toUpperCase()}`;
		}
		
		return firstName.charAt(0).toUpperCase();
	} catch (error) {
		console.error("Error formatting name:", error, "Value:", fullname);
		return 'NA';
	}
}

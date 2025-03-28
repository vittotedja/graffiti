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

export function formatFullName(fullname: string): string {
	const [firstName, lastName] = fullname.split(" ");
	return `${firstName.charAt(0)}${lastName.charAt(0)}`;
}

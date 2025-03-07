import {format, formatDistanceToNow} from "date-fns";

export function formatDate(date: Date | string) {
	const dateObj = typeof date === "string" ? new Date(date) : date;
	return format(dateObj, "MMM d, yyyy");
}

export function formatRelativeTime(date: Date | string) {
	const dateObj = typeof date === "string" ? new Date(date) : date;
	return formatDistanceToNow(dateObj, {addSuffix: true});
}

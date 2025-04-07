package util

import "strings"

func ExtractKeyFromMediaURL(mediaURL string) string {
	parts := strings.SplitN(mediaURL, "/", 4)
	if len(parts) < 4 {
		return ""
	}
	return parts[3] // the path after the domain
}

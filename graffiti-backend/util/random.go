package util

import (
	"fmt"
    "math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Create a global random number generator
var globalRng = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	// Use the global random number generator with a seed
	globalRng.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		idx := globalRng.Intn(k)
		sb.WriteByte(alphabet[idx])
	}
	return sb.String()
}

// RandomUsername generates a random username
func RandomUsername() string {
	return RandomString(6)
}

// RandomEmail generates a random email
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}

// Random Fullname generates a random full name
func RandomFullname() string {
	return fmt.Sprintf("%s %s", RandomString(6), RandomString(6))
}
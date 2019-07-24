package tools

import "os"

// ExistsFile tests if a file exists, or fail if any error occurs
func ExistsFile(path string) bool {
	if _, err := os.Stat(path); err != nil || os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

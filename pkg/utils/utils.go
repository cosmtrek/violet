package utils

import (
	"os"
)

// FileExists checks if files exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

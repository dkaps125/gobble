package utils

import "os"

func DirectoryExists(directory string) bool {
	if _, err := os.Stat(directory); err == nil {
		return true
	} else {
		return !os.IsNotExist(err)
	}
}

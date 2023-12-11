package utils

import "os"

func FileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil || os.IsExist(err) {
		return true
	}
	return false
}

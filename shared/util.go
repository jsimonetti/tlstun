package shared

import (
	"os"
	"time"
)

func PathExists(name string) bool {
	_, err := os.Lstat(name)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

var readTimeout = 10 * time.Second

//go:build !darwin

package api

import (
	"os"
	"time"
)

func getBestEffortCreatedTime(info os.FileInfo) time.Time {
	return info.ModTime()
}

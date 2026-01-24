//go:build darwin

package api

import (
	"os"
	"syscall"
	"time"
)

func getBestEffortCreatedTime(info os.FileInfo) time.Time {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return info.ModTime()
	}
	return time.Unix(int64(stat.Birthtimespec.Sec), int64(stat.Birthtimespec.Nsec))
}

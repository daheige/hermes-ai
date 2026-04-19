package utils

import (
	"time"
)

func GetTimestamp() int64 {
	return time.Now().Unix()
}

// CalcElapsedTime return the elapsed time in milliseconds (ms)
func CalcElapsedTime(start time.Time) int64 {
	return time.Now().Sub(start).Milliseconds()
}

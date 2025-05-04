package core

import (
	"time"
)

func GetTimeFromUnixMicroseconds(mks uint64) time.Time {
	return time.Unix(0, int64(mks*1e3))
}

func GetUnixMicrosecondsFromTime(timeValue time.Time) uint64 {
	return uint64(timeValue.UnixNano() / 1e3)
}

func GetUnixMillisecondsFromTime(timeValue time.Time) int64 {
	return int64(timeValue.UnixNano() / 1e6)
}

func GetUnixSecondsFromTime(timeValue time.Time) int64 {
	return timeValue.Unix()
}

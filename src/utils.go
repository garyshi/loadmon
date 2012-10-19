package main

import "time"

func ToTimestamp(t time.Time) uint32 {
	// seconds from 2000-01-01 00:00:00 UTC, can work till year 2136
	ts := t.Unix() - 946684800
	return uint32(ts)
}

func FromTimestamp(ts uint32) time.Time {
	t := int64(ts) + 946684800
	return time.Unix(t, 0)
}

func GetTimestamp() uint32 {
	t := time.Now()
	return ToTimestamp(t)
}

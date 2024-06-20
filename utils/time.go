package utils

import "time"

const TimeDuration = 20 * time.Second

type alarmToDispatcherTime struct {
}

func NowTime() int64 {
	return time.Now().Unix()
}

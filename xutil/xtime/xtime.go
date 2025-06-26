package xtime

import "time"

// Now 获取当前时间
func Now() time.Time {
	return time.Now()
}

// NowUnix 秒时间戳
func NowUnix() int64 {
	return time.Now().Unix()
}

// NowUnixMilli 毫秒时间戳
func NowUnixMilli() int64 {
	return time.Now().UnixMilli()
}

// NowUnixNano 纳秒时间戳
func NowUnixNano() int64 {
	return time.Now().UnixNano()
}

// NowUnixMicro 微秒时间戳
func NowUnixMicro() int64 {
	return time.Now().UnixMicro()
}

// Sub 计算时间差
func Sub(t time.Time) time.Duration {
	return time.Now().Sub(t)
}

//IsSameDay 判断是否是同一天
func IsSameDay(t1, t2 time.Time) bool {
	return t1.Truncate(24*time.Hour) == t2.Truncate(24*time.Hour)
}

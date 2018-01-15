package bulitin

import "time"

type Time struct{}

func (Time) Now() int {
	return int(time.Now().UnixNano() / 1000000)
}
func (Time) Format(format string) string {
	return time.Now().Format(format)
}
func (Time) Add(n int, format string) int {
	return int(time.Now().Add(time.Duration(n)*time.Second).UnixNano() / 1000000)
}
func (Time) AddDate(y, m, d int, format string) int {
	return int(time.Now().AddDate(y, m, d).UnixNano() / 1000000)
}
func (Time) Parse(number int, format string) string {
	return time.Unix(int64(number/1000), 0).Format(format)
}
func (Time) Duration(i int) time.Duration {
	return time.Duration(i)
}
func (Time) Sleep(n int) {
	if n > 60 {
		panic("sleep时间不能超过60秒")
	}
	time.Sleep(time.Duration(n) * time.Second)
}

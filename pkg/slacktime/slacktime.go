package slacktime

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ParseTime(t time.Time) string {
	s := t.Unix()
	m := t.UnixMicro() - s*1e6
	return fmt.Sprintf("%d.%d", s, m)
}

func ParseString(s string) time.Time {
	parts := strings.Split(s, ".")
	secs, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(err.Error())
	}
	mils, err := strconv.Atoi(parts[1])
	if err != nil {
		panic(err.Error())
	}
	return time.Unix(int64(secs), int64(mils)*1e3)
}

package utility

import (
	"time"

	"github.com/beego/beego/v2/core/logs"
)

func TimeStartOfDay() (resultTime time.Time, timeresult string) {
	// Get the current time
	now := time.Now()
	// Create a new time.Time value with the same year, month, and day, but time set to 00:00:00
	resultTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	timeresult = resultTime.Format(time.DateTime)
	return
}

func TimeEndOfDay() (resultTime time.Time, timeresult string) {
	// Get the current time
	now := time.Now()
	// Create a new time.Time value with the same year, month, and day, but time set to 23:59:59.999999999
	resultTime = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Nanosecond-1), now.Location())
	timeresult = resultTime.Format(time.DateTime)
	return
}

func TimeStringBetween(startTime, endTime string) (durationDay int, err error) {
	var startTimetmp time.Time

	logs.Info(startTime)
	if len(startTime) == 19 {
		startTimetmp, err = time.Parse(time.DateOnly, startTime)
	} else {
		startTimetmp, err = time.Parse(time.DateOnly, startTime)
	}
	if err != nil {
		logs.Error("[TimeStringBetween] Parse Starttime error", err)
	}

	var endTimetmp time.Time

	logs.Info(endTime)
	if len(endTime) == 19 {
		endTimetmp, err = time.Parse(time.DateOnly, endTime)
	} else {
		endTimetmp, err = time.Parse(time.DateOnly, endTime)
	}
	if err != nil {
		logs.Error("[TimeStringBetween] Parse EnndTime error", err)
		return
	}

	durationDay = int(endTimetmp.Sub(startTimetmp).Hours() / 24)
	return
}

func TimeParseWithoutError(timeString string, layout string) (result uint64) {
	tmp, _ := time.Parse(layout, timeString)
	result = uint64(tmp.Unix())
	return
}

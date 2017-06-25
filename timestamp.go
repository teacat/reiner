package reiner

import "strings"

type Timestamp struct {
	value interface{}
	query string
}

func (t Timestamp) IsDate(date string) Timestamp {
	t.query = "DATE(FROM_UNIXTIME(%s)) = %s "
	t.value = date
	return t
}

func (t Timestamp) IsYear(year int) Timestamp {
	t.query = "YEAR(FROM_UNIXTIME(%s)) = %s "
	t.value = year
	return t
}

func (t Timestamp) IsMonth(month interface{}) Timestamp {
	t.query = "MONTH(FROM_UNIXTIME(%s)) = %s "
	switch v := month.(type) {
	case int:
		t.value = v
	case string:
		list := map[string]int{
			"january":   1,
			"february":  2,
			"march":     3,
			"april":     4,
			"may":       5,
			"june":      6,
			"july":      7,
			"august":    8,
			"september": 9,
			"october":   10,
			"november":  11,
			"december":  12,
		}
		t.value = list[strings.ToLower(v)]
	}
	return t
}

func (t Timestamp) IsDay(day int) Timestamp {
	t.query = "DAY(FROM_UNIXTIME(%s)) = %s "
	t.value = day
	return t
}

func (t Timestamp) IsWeekday(weekday interface{}) Timestamp {
	t.query = "WEEKDAY(FROM_UNIXTIME(%s)) = %s "
	switch v := weekday.(type) {
	case int:
		t.value = v
	case string:
		list := map[string]int{
			"monday":    1,
			"tuesday":   2,
			"wednesday": 3,
			"thursday":  4,
			"friday":    5,
			"saturday":  6,
			"sunday":    7,
		}
		t.value = list[strings.ToLower(v)]
	}
	return t
}

func (t Timestamp) IsHour(hour int) Timestamp {
	t.query = "HOUR(FROM_UNIXTIME(%s)) = %s "
	t.value = hour
	return t
}

func (t Timestamp) IsMinute(minute int) Timestamp {
	t.query = "MINUTE(FROM_UNIXTIME(%s)) = %s "
	t.value = minute
	return t
}

func (t Timestamp) IsSecond(second int) Timestamp {
	t.query = "SECOND(FROM_UNIXTIME(%s)) = %s "
	t.value = second
	return t
}

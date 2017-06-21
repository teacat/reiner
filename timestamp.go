package reiner

type Timestamp struct {
}

type timestampValue struct {
}

func (t *Timestamp) Now(format string) (v timestampValue) {
	return
}

func (t *Timestamp) IsDate(date string) (v timestampValue) {
	return
}

func (t *Timestamp) IsYear(year int) (v timestampValue) {
	return
}

func (t *Timestamp) IsMonth(month interface{}) (v timestampValue) {
	return
}

func (t *Timestamp) IsDay(day int) (v timestampValue) {
	return
}

func (t *Timestamp) IsWeekday(weekday interface{}) (v timestampValue) {
	return
}

func (t *Timestamp) IsHour(hour int) (v timestampValue) {
	return
}

func (t *Timestamp) IsMinute(minute int) (v timestampValue) {
	return
}

func (t *Timestamp) IsSecond(second int) (v timestampValue) {
	return
}

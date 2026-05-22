package core

import (
	"math"
	"time"
)

type DateLocator struct {
	Location *time.Location
}

func (l DateLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	if math.IsNaN(minVal) || math.IsNaN(maxVal) || math.IsInf(minVal, 0) || math.IsInf(maxVal, 0) {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	if targetCount <= 0 {
		targetCount = 6
	}

	minTime := dateNumberToTime(minVal, l.location())
	maxTime := dateNumberToTime(maxVal, l.location())
	if !maxTime.After(minTime) {
		return []float64{minVal}
	}

	interval := chooseDateTickInterval(minTime, maxTime, targetCount)
	current := interval.align(minTime)
	if current.Before(minTime) {
		current = interval.next(current)
	}

	ticks := make([]float64, 0, targetCount+2)
	guard := targetCount*4 + 16
	for i := 0; i < guard && !current.After(maxTime); i++ {
		ticks = append(ticks, timeToDateNumber(current))
		current = interval.next(current)
	}

	if len(ticks) == 0 {
		ticks = append(ticks, minVal, maxVal)
	}
	return dedupeTicks(ticks)
}

func (l DateLocator) location() *time.Location {
	if l.Location != nil {
		return l.Location
	}
	return time.UTC
}

type DayLocator struct {
	ByMonthDay []int
	Interval   int
	Location   *time.Location
}

func (l DayLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	if math.IsNaN(minVal) || math.IsNaN(maxVal) || math.IsInf(minVal, 0) || math.IsInf(maxVal, 0) {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}

	loc := l.location()
	minTime := dateNumberToTime(minVal, loc)
	maxTime := dateNumberToTime(maxVal, loc)
	if !maxTime.After(minTime) {
		return []float64{minVal}
	}

	interval := l.Interval
	if interval <= 0 {
		interval = 1
	}
	monthDays := validMonthDays(l.ByMonthDay)
	current := time.Date(minTime.Year(), minTime.Month(), minTime.Day(), 0, 0, 0, 0, loc)
	if current.Before(minTime) {
		current = current.AddDate(0, 0, 1)
	}

	guard := int(maxTime.Sub(minTime).Hours()/24) + 370
	ticks := make([]float64, 0, targetCount+2)
	startOrdinal := current.YearDay()
	for i := 0; i < guard && !current.After(maxTime); i++ {
		if len(monthDays) > 0 {
			if monthDays[current.Day()] {
				ticks = append(ticks, timeToDateNumber(current))
			}
		} else if (current.YearDay()-startOrdinal)%interval == 0 {
			ticks = append(ticks, timeToDateNumber(current))
		}
		current = current.AddDate(0, 0, 1)
	}
	return dedupeTicks(ticks)
}

func (l DayLocator) location() *time.Location {
	if l.Location != nil {
		return l.Location
	}
	return time.UTC
}

func validMonthDays(days []int) map[int]bool {
	if len(days) == 0 {
		return nil
	}
	out := make(map[int]bool, len(days))
	for _, day := range days {
		if day >= 1 && day <= 31 {
			out[day] = true
		}
	}
	return out
}

type AutoDateFormatter struct {
	Min      float64
	Max      float64
	Location *time.Location
}

func (f AutoDateFormatter) Format(x float64) string {
	layout := chooseDateLabelLayout(f.Min, f.Max)
	return DateFormatter{Layout: layout, Location: f.location()}.Format(x)
}

func (f AutoDateFormatter) location() *time.Location {
	if f.Location != nil {
		return f.Location
	}
	return time.UTC
}

type DateFormatter struct {
	Layout   string
	Location *time.Location
}

func (f DateFormatter) Format(x float64) string {
	layout := f.Layout
	if layout == "" {
		layout = time.RFC3339
	}
	return dateNumberToTime(x, f.location()).Format(layout)
}

func (f DateFormatter) location() *time.Location {
	if f.Location != nil {
		return f.Location
	}
	return time.UTC
}

type dateTickInterval struct {
	unit string
	step int
}

func chooseDateTickInterval(minTime, maxTime time.Time, targetCount int) dateTickInterval {
	if !maxTime.After(minTime) {
		return dateTickInterval{unit: "day", step: 1}
	}

	// Match Matplotlib's AutoDateLocator selection model: first choose the
	// finest frequency that can produce at least minticks, then choose the
	// first interval that stays below that frequency's maxticks budget.
	// targetCount is intentionally not treated as a hard cap because
	// Matplotlib's date locator will allow dense daily labels in compact axes.
	const minticks = 5
	candidates := []struct {
		interval dateTickInterval
		count    int
		maxticks int
	}{
		{dateTickInterval{unit: "year"}, dateSpanYears(minTime, maxTime), 11},
		{dateTickInterval{unit: "month"}, dateSpanMonths(minTime, maxTime), 12},
		{dateTickInterval{unit: "day"}, int(maxTime.Sub(minTime).Hours() / 24), 11},
		{dateTickInterval{unit: "hour"}, int(maxTime.Sub(minTime).Hours()), 12},
		{dateTickInterval{unit: "minute"}, int(maxTime.Sub(minTime).Minutes()), 11},
		{dateTickInterval{unit: "second"}, int(maxTime.Sub(minTime).Seconds()), 11},
	}

	intervals := map[string][]int{
		"year":   {1, 2, 4, 5, 10, 20, 40, 50, 100, 200, 400, 500, 1000, 2000, 4000, 5000, 10000},
		"month":  {1, 2, 3, 4, 6},
		"day":    {1, 2, 4, 7, 14},
		"hour":   {1, 2, 3, 4, 6, 12},
		"minute": {1, 5, 10, 15, 30},
		"second": {1, 5, 10, 15, 30},
	}

	for _, candidate := range candidates {
		if candidate.count < minticks {
			continue
		}
		steps := intervals[candidate.interval.unit]
		for _, step := range steps {
			if candidate.count <= step*(candidate.maxticks-1) {
				candidate.interval.step = step
				return candidate.interval
			}
		}
		candidate.interval.step = steps[len(steps)-1]
		return candidate.interval
	}

	span := maxTime.Sub(minTime)
	if span >= time.Second {
		return dateTickInterval{unit: "second", step: 1}
	}
	return dateTickInterval{unit: "second", step: 1}
}

func dateSpanYears(minTime, maxTime time.Time) int {
	years := maxTime.Year() - minTime.Year()
	if maxTime.YearDay() < minTime.YearDay() {
		years--
	}
	if years < 0 {
		return 0
	}
	return years
}

func dateSpanMonths(minTime, maxTime time.Time) int {
	months := (maxTime.Year()-minTime.Year())*12 + int(maxTime.Month()) - int(minTime.Month())
	if maxTime.Day() < minTime.Day() {
		months--
	}
	if months < 0 {
		return 0
	}
	return months
}

func (i dateTickInterval) align(t time.Time) time.Time {
	switch i.unit {
	case "year":
		year := (t.Year() / i.step) * i.step
		return time.Date(year, time.January, 1, 0, 0, 0, 0, t.Location())
	case "month":
		monthIndex := int(t.Month()) - 1
		aligned := (monthIndex / i.step) * i.step
		return time.Date(t.Year(), time.Month(aligned+1), 1, 0, 0, 0, 0, t.Location())
	case "day":
		y, m, d := t.Date()
		d = ((d - 1) / i.step * i.step) + 1
		return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
	case "hour":
		y, m, d := t.Date()
		hour := (t.Hour() / i.step) * i.step
		return time.Date(y, m, d, hour, 0, 0, 0, t.Location())
	case "minute":
		y, m, d := t.Date()
		minute := (t.Minute() / i.step) * i.step
		return time.Date(y, m, d, t.Hour(), minute, 0, 0, t.Location())
	default:
		y, m, d := t.Date()
		second := (t.Second() / i.step) * i.step
		return time.Date(y, m, d, t.Hour(), t.Minute(), second, 0, t.Location())
	}
}

func (i dateTickInterval) next(t time.Time) time.Time {
	switch i.unit {
	case "year":
		return t.AddDate(i.step, 0, 0)
	case "month":
		return t.AddDate(0, i.step, 0)
	case "day":
		return t.AddDate(0, 0, i.step)
	case "hour":
		return t.Add(time.Duration(i.step) * time.Hour)
	case "minute":
		return t.Add(time.Duration(i.step) * time.Minute)
	default:
		return t.Add(time.Duration(i.step) * time.Second)
	}
}

func chooseDateLabelLayout(minVal, maxVal float64) string {
	span := math.Abs(maxVal - minVal)
	switch {
	case span >= 2*365*24*3600:
		return "2006"
	case span >= 90*24*3600:
		return "Jan 2006"
	case span >= 2*24*3600:
		return "2006-01-02"
	case span >= 24*3600:
		return "2006-01-02 15:04"
	case span >= 60:
		return "15:04"
	default:
		return "15:04:05"
	}
}

func timeToDateNumber(t time.Time) float64 {
	t = t.UTC()
	return float64(t.Unix()) + float64(t.Nanosecond())/1e9
}

func dateNumberToTime(v float64, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	sec, frac := math.Modf(v)
	nsec := int64(math.Round(frac * 1e9))
	return time.Unix(int64(sec), nsec).In(loc)
}

package core

import (
	"math"
	"strings"
	"sync/atomic"
	"time"
)

type DateLocator struct {
	Location *time.Location
}

func (l DateLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	if math.IsNaN(minVal) || math.IsNaN(maxVal) || math.IsInf(minVal, 0) || math.IsInf(maxVal, 0) {
		return nil
	}
	minVal, maxVal = dateNonsingular(minVal, maxVal)
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

type YearLocator struct {
	Base     int
	Month    time.Month
	Day      int
	Location *time.Location
}

func (l YearLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
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

	base := l.Base
	if base <= 0 {
		base = 1
	}
	month := l.Month
	if month < time.January || month > time.December {
		month = time.January
	}
	day := l.Day
	if day <= 0 {
		day = 1
	}
	year := (minTime.Year() / base) * base
	current := safeDate(year, month, day, loc)
	for current.Before(minTime) {
		current = safeDate(current.Year()+base, month, day, loc)
	}

	guard := dateSpanYears(minTime, maxTime)/base + targetCount + 8
	if guard < 16 {
		guard = 16
	}
	ticks := make([]float64, 0, targetCount+2)
	for i := 0; i < guard && !current.After(maxTime); i++ {
		ticks = append(ticks, timeToDateNumber(current))
		current = safeDate(current.Year()+base, month, day, loc)
	}
	return dedupeTicks(ticks)
}

func (l YearLocator) location() *time.Location {
	if l.Location != nil {
		return l.Location
	}
	return time.UTC
}

type MonthLocator struct {
	ByMonth    []time.Month
	ByMonthDay int
	Interval   int
	Location   *time.Location
}

func (l MonthLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
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
	day := l.ByMonthDay
	if day <= 0 {
		day = 1
	}
	months := validMonths(l.ByMonth)
	current := time.Date(minTime.Year(), minTime.Month(), 1, 0, 0, 0, 0, loc)
	if current.Before(minTime) {
		current = current.AddDate(0, 1, 0)
	}
	startIndex := current.Year()*12 + int(current.Month()) - 1
	guard := dateSpanMonths(minTime, maxTime) + targetCount + 16
	if guard < 24 {
		guard = 24
	}
	ticks := make([]float64, 0, targetCount+2)
	for i := 0; i < guard && !current.After(maxTime); i++ {
		monthIndex := current.Year()*12 + int(current.Month()) - 1
		if (monthIndex-startIndex)%interval == 0 && (len(months) == 0 || months[current.Month()]) {
			tick := safeDate(current.Year(), current.Month(), day, loc)
			if !tick.Before(minTime) && !tick.After(maxTime) {
				ticks = append(ticks, timeToDateNumber(tick))
			}
		}
		current = current.AddDate(0, 1, 0)
	}
	return dedupeTicks(ticks)
}

func (l MonthLocator) location() *time.Location {
	if l.Location != nil {
		return l.Location
	}
	return time.UTC
}

type WeekdayLocator struct {
	ByWeekday []time.Weekday
	Interval  int
	Location  *time.Location
}

func (l WeekdayLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
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
	weekdays := validWeekdays(l.ByWeekday)
	current := time.Date(minTime.Year(), minTime.Month(), minTime.Day(), 0, 0, 0, 0, loc)
	if current.Before(minTime) {
		current = current.AddDate(0, 0, 1)
	}
	startISOYear, startISOWeek := current.ISOWeek()
	guard := int(maxTime.Sub(minTime).Hours()/24) + 14
	ticks := make([]float64, 0, targetCount+2)
	for i := 0; i < guard && !current.After(maxTime); i++ {
		if weekdays[current.Weekday()] {
			isoYear, isoWeek := current.ISOWeek()
			if weeksBetween(startISOYear, startISOWeek, isoYear, isoWeek)%interval == 0 {
				ticks = append(ticks, timeToDateNumber(current))
			}
		}
		current = current.AddDate(0, 0, 1)
	}
	return dedupeTicks(ticks)
}

func (l WeekdayLocator) location() *time.Location {
	if l.Location != nil {
		return l.Location
	}
	return time.UTC
}

type HourLocator struct {
	ByHour   []int
	Interval int
	Location *time.Location
}

func (l HourLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	return clockTicks(minVal, maxVal, l.location(), "hour", validClockValues(l.ByHour, 0, 23), l.Interval, targetCount)
}

func (l HourLocator) location() *time.Location {
	if l.Location != nil {
		return l.Location
	}
	return time.UTC
}

type MinuteLocator struct {
	ByMinute []int
	Interval int
	Location *time.Location
}

func (l MinuteLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	return clockTicks(minVal, maxVal, l.location(), "minute", validClockValues(l.ByMinute, 0, 59), l.Interval, targetCount)
}

func (l MinuteLocator) location() *time.Location {
	if l.Location != nil {
		return l.Location
	}
	return time.UTC
}

type SecondLocator struct {
	BySecond []int
	Interval int
	Location *time.Location
}

func (l SecondLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	return clockTicks(minVal, maxVal, l.location(), "second", validClockValues(l.BySecond, 0, 59), l.Interval, targetCount)
}

func (l SecondLocator) location() *time.Location {
	if l.Location != nil {
		return l.Location
	}
	return time.UTC
}

type MicrosecondLocator struct {
	Interval int
	Location *time.Location
}

func (l MicrosecondLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	if math.IsNaN(minVal) || math.IsNaN(maxVal) || math.IsInf(minVal, 0) || math.IsInf(maxVal, 0) {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	if minVal == maxVal {
		return []float64{minVal}
	}
	interval := l.Interval
	if interval <= 0 {
		interval = 1
	}

	minUs := int64(math.Ceil(minVal * microsecondsPerDay))
	maxUs := int64(math.Floor(maxVal * microsecondsPerDay))
	step := int64(interval)
	start := ceilDivInt64(minUs, step) * step
	if start > maxUs {
		return nil
	}

	estimated := int((maxUs-start)/step) + 1
	guard := targetCount*4 + 16
	if guard < 32 {
		guard = 32
	}
	if estimated > guard {
		estimated = guard
	}
	ticks := make([]float64, 0, estimated)
	for value, i := start, 0; value <= maxUs && i < guard; value, i = value+step, i+1 {
		ticks = append(ticks, float64(value)/microsecondsPerDay)
	}
	return dedupeTicks(ticks)
}

func (l MicrosecondLocator) location() *time.Location {
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

func validMonths(months []time.Month) map[time.Month]bool {
	if len(months) == 0 {
		return nil
	}
	out := make(map[time.Month]bool, len(months))
	for _, month := range months {
		if month >= time.January && month <= time.December {
			out[month] = true
		}
	}
	return out
}

func validWeekdays(days []time.Weekday) map[time.Weekday]bool {
	out := make(map[time.Weekday]bool, len(days))
	if len(days) == 0 {
		for day := time.Sunday; day <= time.Saturday; day++ {
			out[day] = true
		}
		return out
	}
	for _, day := range days {
		if day >= time.Sunday && day <= time.Saturday {
			out[day] = true
		}
	}
	return out
}

func validClockValues(values []int, minVal, maxVal int) map[int]bool {
	if len(values) == 0 {
		return nil
	}
	out := make(map[int]bool, len(values))
	for _, value := range values {
		if value >= minVal && value <= maxVal {
			out[value] = true
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

type ConciseDateFormatter struct {
	Location    *time.Location
	Formats     []string
	ZeroFormats []string
}

func (f ConciseDateFormatter) Format(x float64) string {
	return dateNumberToTime(x, f.location()).Format(f.formatForLevel(0))
}

func (f ConciseDateFormatter) FormatTick(x float64, index int, ticks []float64) string {
	if len(ticks) == 0 {
		return f.Format(x)
	}
	current := dateNumberToTime(x, f.location())
	level := conciseDateCommonLevel(ticks, f.location())
	layout := f.formatForLevel(level)
	if zeroLayout := f.zeroFormatForLevel(level); conciseDateIsZero(current, level) && zeroLayout != "" {
		layout = zeroLayout
	}
	label := current.Format(layout)
	if level >= 5 {
		label = trimConciseSubseconds(label)
	}
	return label
}

func (f ConciseDateFormatter) OffsetText(ticks []float64) string {
	if len(ticks) == 0 {
		return ""
	}
	level := conciseDateCommonLevel(ticks, f.location())
	layout := f.offsetFormatForLevel(level)
	if layout == "" {
		return ""
	}
	return dateNumberToTime(ticks[len(ticks)-1], f.location()).Format(layout)
}

func (f ConciseDateFormatter) location() *time.Location {
	if f.Location != nil {
		return f.Location
	}
	return time.UTC
}

func (f ConciseDateFormatter) formatForLevel(level int) string {
	defaults := []string{"2006", "Jan", "02", "15:04", "15:04", "05.000000"}
	if level >= 0 && level < len(f.Formats) && f.Formats[level] != "" {
		return f.Formats[level]
	}
	if level < 0 {
		level = 0
	}
	if level >= len(defaults) {
		level = len(defaults) - 1
	}
	return defaults[level]
}

func (f ConciseDateFormatter) zeroFormatForLevel(level int) string {
	defaults := []string{"", "2006", "Jan", "Jan-02", "15:04", "15:04"}
	if level < 0 {
		level = 0
	}
	if level >= len(defaults) {
		level = len(defaults) - 1
	}
	if level >= 0 && level < len(f.ZeroFormats) && f.ZeroFormats[level] != "" {
		return f.ZeroFormats[level]
	}
	if len(f.Formats) > 0 && level > 0 && level-1 < len(f.Formats) && f.Formats[level-1] != "" {
		return f.Formats[level-1]
	}
	return defaults[level]
}

func (f ConciseDateFormatter) offsetFormatForLevel(level int) string {
	defaults := []string{"", "2006", "2006-Jan", "2006-Jan-02", "2006-Jan-02", "2006-Jan-02 15:04"}
	if level < 0 {
		level = 0
	}
	if level >= len(defaults) {
		level = len(defaults) - 1
	}
	return defaults[level]
}

func conciseDateCommonLevel(ticks []float64, loc *time.Location) int {
	if len(ticks) == 0 {
		return 0
	}
	for level := 5; level >= 0; level-- {
		first := conciseDateField(dateNumberToTime(ticks[0], loc), level)
		for _, tick := range ticks[1:] {
			if conciseDateField(dateNumberToTime(tick, loc), level) != first {
				return level
			}
		}
	}
	return 5
}

func conciseDateField(t time.Time, level int) int {
	switch level {
	case 0:
		return t.Year()
	case 1:
		return int(t.Month())
	case 2:
		return t.Day()
	case 3:
		return t.Hour()
	case 4:
		return t.Minute()
	default:
		return t.Second()
	}
}

func conciseDateIsZero(t time.Time, level int) bool {
	switch {
	case level == 1:
		return t.Month() == time.January
	case level == 2:
		return t.Day() == 1
	case level == 3:
		return t.Hour() == 0
	case level == 4:
		return t.Minute() == 0
	case level >= 5:
		return t.Second() == 0 && t.Nanosecond() == 0
	default:
		return false
	}
}

func trimConciseSubseconds(label string) string {
	if !strings.Contains(label, ".") {
		return label
	}
	return strings.TrimRight(strings.TrimRight(label, "0"), ".")
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
		{dateTickInterval{unit: "microsecond"}, int(maxTime.Sub(minTime) / time.Microsecond), 8},
	}

	intervals := map[string][]int{
		"year":        {1, 2, 4, 5, 10, 20, 40, 50, 100, 200, 400, 500, 1000, 2000, 4000, 5000, 10000},
		"month":       {1, 2, 3, 4, 6},
		"day":         {1, 2, 4, 7, 14},
		"hour":        {1, 2, 3, 4, 6, 12},
		"minute":      {1, 5, 10, 15, 30},
		"second":      {1, 5, 10, 15, 30},
		"microsecond": {1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000, 20000, 50000, 100000, 200000, 500000, 1000000},
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

	return dateTickInterval{unit: "microsecond", step: 1}
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
	case "second":
		y, m, d := t.Date()
		second := (t.Second() / i.step) * i.step
		return time.Date(y, m, d, t.Hour(), t.Minute(), second, 0, t.Location())
	default:
		y, m, d := t.Date()
		us := (t.Nanosecond() / int(time.Microsecond)) / i.step * i.step
		return time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), us*int(time.Microsecond), t.Location())
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
	case "second":
		return t.Add(time.Duration(i.step) * time.Second)
	default:
		return t.Add(time.Duration(i.step) * time.Microsecond)
	}
}

func ceilDivInt64(n, d int64) int64 {
	q := n / d
	r := n % d
	if r != 0 && ((r > 0) == (d > 0)) {
		q++
	}
	return q
}

func clockTicks(minVal, maxVal float64, loc *time.Location, unit string, allowed map[int]bool, interval, targetCount int) []float64 {
	if math.IsNaN(minVal) || math.IsNaN(maxVal) || math.IsInf(minVal, 0) || math.IsInf(maxVal, 0) {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	minTime := dateNumberToTime(minVal, loc)
	maxTime := dateNumberToTime(maxVal, loc)
	if !maxTime.After(minTime) {
		return []float64{minVal}
	}
	if interval <= 0 {
		interval = 1
	}
	current := truncateDateTime(minTime, unit)
	if current.Before(minTime) {
		current = addDateUnit(current, unit, 1)
	}
	startOrdinal := clockOrdinal(current, unit)
	guard := clockGuard(minTime, maxTime, unit) + targetCount + 16
	ticks := make([]float64, 0, targetCount+2)
	for i := 0; i < guard && !current.After(maxTime); i++ {
		if (clockOrdinal(current, unit)-startOrdinal)%interval == 0 && clockValueAllowed(current, unit, allowed) {
			ticks = append(ticks, timeToDateNumber(current))
		}
		current = addDateUnit(current, unit, 1)
	}
	return dedupeTicks(ticks)
}

func truncateDateTime(t time.Time, unit string) time.Time {
	y, m, d := t.Date()
	switch unit {
	case "hour":
		return time.Date(y, m, d, t.Hour(), 0, 0, 0, t.Location())
	case "minute":
		return time.Date(y, m, d, t.Hour(), t.Minute(), 0, 0, t.Location())
	default:
		return time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), 0, t.Location())
	}
}

func addDateUnit(t time.Time, unit string, step int) time.Time {
	switch unit {
	case "hour":
		return t.Add(time.Duration(step) * time.Hour)
	case "minute":
		return t.Add(time.Duration(step) * time.Minute)
	default:
		return t.Add(time.Duration(step) * time.Second)
	}
}

func clockOrdinal(t time.Time, unit string) int {
	switch unit {
	case "hour":
		return int(t.Unix() / int64(time.Hour/time.Second))
	case "minute":
		return int(t.Unix() / int64(time.Minute/time.Second))
	default:
		return int(t.Unix())
	}
}

func clockGuard(minTime, maxTime time.Time, unit string) int {
	switch unit {
	case "hour":
		return int(maxTime.Sub(minTime).Hours()) + 4
	case "minute":
		return int(maxTime.Sub(minTime).Minutes()) + 4
	default:
		return int(maxTime.Sub(minTime).Seconds()) + 4
	}
}

func clockValueAllowed(t time.Time, unit string, allowed map[int]bool) bool {
	if len(allowed) == 0 {
		return true
	}
	switch unit {
	case "hour":
		return allowed[t.Hour()]
	case "minute":
		return allowed[t.Minute()]
	default:
		return allowed[t.Second()]
	}
}

func safeDate(year int, month time.Month, day int, loc *time.Location) time.Time {
	if day <= 0 {
		day = 1
	}
	last := daysInMonth(year, month, loc)
	if day > last {
		day = last
	}
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

func daysInMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

func weeksBetween(startYear, startWeek, endYear, endWeek int) int {
	start := isoWeekStart(startYear, startWeek)
	end := isoWeekStart(endYear, endWeek)
	return int(end.Sub(start).Hours() / (24 * 7))
}

func isoWeekStart(year, week int) time.Time {
	jan4 := time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC)
	offset := (int(jan4.Weekday()) + 6) % 7
	week1 := jan4.AddDate(0, 0, -offset)
	return week1.AddDate(0, 0, (week-1)*7)
}

func chooseDateLabelLayout(minVal, maxVal float64) string {
	// span is in date-number units, i.e. days since the epoch.
	span := math.Abs(maxVal - minVal)
	const (
		oneSecond = 1.0 / 86400.0
		oneMinute = 60.0 * oneSecond
	)
	switch {
	case span >= 2*365: // ~2 years
		return "2006"
	case span >= 90: // 90 days
		return "Jan 2006"
	case span >= 2: // 2 days
		return "2006-01-02"
	case span >= 1: // 1 day
		return "2006-01-02 15:04"
	case span >= oneMinute:
		return "15:04"
	default:
		return "15:04:05"
	}
}

// dateNonsingular expands a degenerate date range, mirroring
// AutoDateLocator.nonsingular: non-finite ranges fall back to a one-day window
// and an exactly-singular range expands to roughly a four-year period.
func dateNonsingular(minVal, maxVal float64) (float64, float64) {
	const daysPerYear = 365.2425
	if math.IsNaN(minVal) || math.IsNaN(maxVal) || math.IsInf(minVal, 0) || math.IsInf(maxVal, 0) {
		return 0, 1 // 1970-01-01 .. 1970-01-02
	}
	if maxVal < minVal {
		minVal, maxVal = maxVal, minVal
	}
	if minVal == maxVal {
		minVal -= daysPerYear * 2
		maxVal += daysPerYear * 2
	}
	return minVal, maxVal
}

// Date numbers follow Matplotlib's convention: floating-point days since a
// configurable epoch (default 1970-01-01T00:00:00Z). See core/dates.go for the
// public Date2Num / Num2Date / SetEpoch / GetEpoch surface.
var dateEpoch = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

// dateEpochUsed is set once any conversion happens; it locks SetEpoch. It is
// atomic so concurrent conversions (or a conversion racing SetEpoch) are
// data-race free under the Go race detector.
var dateEpochUsed atomic.Bool

const (
	nanosPerDay        = 24.0 * 60.0 * 60.0 * 1e9
	microsecondsPerDay = 24.0 * 60.0 * 60.0 * 1e6
	secondsPerDay      = 24.0 * 60.0 * 60.0

	// maxDurationMicros / minDurationMicros bound the microsecond offset for
	// which micros*1000 still fits a time.Duration (int64 nanoseconds). Beyond
	// this the time.Duration paths below saturate (~292 years from the epoch).
	maxDurationMicros = math.MaxInt64 / 1000
	minDurationMicros = math.MinInt64 / 1000
)

func timeToDateNumber(t time.Time) float64 {
	dateEpochUsed.Store(true)
	t = t.UTC()
	// Common (in-range) path: identical to the historical computation, which the
	// date-parity fixtures are calibrated against. time.Time.Sub saturates beyond
	// ~292 years from the epoch, so for those extreme offsets fall back to
	// time.Duration-free arithmetic via Unix seconds (a full-range int64 count).
	if d := t.Sub(dateEpoch); d != time.Duration(math.MaxInt64) && d != time.Duration(math.MinInt64) {
		return float64(d) / nanosPerDay
	}
	secs := t.Unix() - dateEpoch.Unix()
	fracNanos := int64(t.Nanosecond()) - int64(dateEpoch.Nanosecond())
	return float64(secs)/secondsPerDay + float64(fracNanos)/nanosPerDay
}

func dateNumberToTime(v float64, loc *time.Location) time.Time {
	dateEpochUsed.Store(true)
	if loc == nil {
		loc = time.UTC
	}
	// Round to the nearest microsecond, like Matplotlib's num2date: fractional
	// days are not exactly representable in float64, so nanosecond rounding
	// leaves a sub-microsecond drift (e.g. 02:40:00 -> 02:39:59.9999997).
	micros := int64(math.Round(v * microsecondsPerDay))
	// In-range path is identical to before; large offsets would overflow
	// time.Duration(micros)*time.Microsecond, so add them via Unix seconds.
	if micros > maxDurationMicros || micros < minDurationMicros {
		secs := micros / 1_000_000
		rem := micros % 1_000_000
		return time.Unix(dateEpoch.Unix()+secs, int64(dateEpoch.Nanosecond())+rem*1000).In(loc)
	}
	return dateEpoch.Add(time.Duration(micros) * time.Microsecond).In(loc)
}

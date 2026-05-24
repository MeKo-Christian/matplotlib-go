package examplecatalog

// DateSurfaceAudit records how upstream matplotlib.dates surfaces map to the
// Go date/unit axis implementation. dates.py is tracked as supporting Phase 12
// coverage rather than as a Phase 11 public-surface inventory module.
type DateSurfaceAudit struct {
	UpstreamID string
	Status     PublicSurfaceParityStatus
	GoFiles    []string
	Note       string
}

var dateSurfaceAuditRows = []DateSurfaceAudit{
	{
		UpstreamID: "dates.py:function:date2num",
		Status:     PublicSurfaceDirectEquivalent,
		GoFiles:    []string{"core/units.go"},
		Note:       "Go converts time.Time values to Matplotlib date numbers when unit conversion configures date axes.",
	},
	{
		UpstreamID: "dates.py:function:num2date",
		Status:     PublicSurfaceDirectEquivalent,
		GoFiles:    []string{"core/units.go", "core/date_tick.go"},
		Note:       "Go converts Matplotlib date numbers back to time.Time in UTC or a caller-provided location for date locators and formatters.",
	},
	{
		UpstreamID: "dates.py:class:DateFormatter",
		Status:     PublicSurfaceIdiomaticEquivalent,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Go DateFormatter uses Go time layouts instead of strftime format strings.",
	},
	{
		UpstreamID: "dates.py:class:AutoDateFormatter",
		Status:     PublicSurfacePartial,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Go AutoDateFormatter chooses date/time layouts from the visible span; full matplotlib scale-dictionary customization is not mirrored.",
	},
	{
		UpstreamID: "dates.py:class:ConciseDateFormatter",
		Status:     PublicSurfacePartial,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Go ConciseDateFormatter suppresses repeated date levels across tick sequences; exact offset-string and format-array customization remains partial.",
	},
	{
		UpstreamID: "dates.py:class:DateLocator",
		Status:     PublicSurfacePartial,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Go DateLocator implements AutoDateLocator-style interval selection for common year-to-microsecond spans.",
	},
	{
		UpstreamID: "dates.py:class:AutoDateLocator",
		Status:     PublicSurfacePartial,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Go DateLocator folds AutoDateLocator behavior into the default date locator rather than exposing a separate class.",
	},
	{
		UpstreamID: "dates.py:class:RRuleLocator",
		Status:     PublicSurfaceIntentionalOmission,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Arbitrary dateutil rrule scheduling is intentionally omitted; Go exposes focused Year/Month/Weekday/Day/Hour/Minute/Second/Microsecond locators.",
	},
	{
		UpstreamID: "dates.py:class:YearLocator",
		Status:     PublicSurfaceDirectEquivalent,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Go YearLocator supports base, month, day, and location options.",
	},
	{
		UpstreamID: "dates.py:class:MonthLocator",
		Status:     PublicSurfaceDirectEquivalent,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Go MonthLocator supports selected months, month day, interval, and location options.",
	},
	{
		UpstreamID: "dates.py:class:WeekdayLocator",
		Status:     PublicSurfaceDirectEquivalent,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Go WeekdayLocator supports selected weekdays, interval, and location options.",
	},
	{
		UpstreamID: "dates.py:class:DayLocator",
		Status:     PublicSurfaceDirectEquivalent,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Go DayLocator supports selected month days, interval, and location options.",
	},
	{
		UpstreamID: "dates.py:class:HourLocator",
		Status:     PublicSurfaceDirectEquivalent,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Go HourLocator supports selected hours, interval, and location options.",
	},
	{
		UpstreamID: "dates.py:class:MinuteLocator",
		Status:     PublicSurfaceDirectEquivalent,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Go MinuteLocator supports selected minutes, interval, and location options.",
	},
	{
		UpstreamID: "dates.py:class:SecondLocator",
		Status:     PublicSurfaceDirectEquivalent,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Go SecondLocator supports selected seconds, interval, and location options.",
	},
	{
		UpstreamID: "dates.py:class:MicrosecondLocator",
		Status:     PublicSurfaceDirectEquivalent,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "Go MicrosecondLocator supports microsecond intervals for subsecond ranges.",
	},
	{
		UpstreamID: "dates.py:class:DateConverter",
		Status:     PublicSurfaceIdiomaticEquivalent,
		GoFiles:    []string{"core/units.go", "core/date_tick.go"},
		Note:       "Go unit conversion configures date locators and formatters from time.Time data without exposing a converter class hierarchy.",
	},
	{
		UpstreamID: "dates.py:class:ConciseDateConverter",
		Status:     PublicSurfaceIdiomaticEquivalent,
		GoFiles:    []string{"core/units.go", "core/date_tick.go"},
		Note:       "Concise date formatting is exposed through ConciseDateFormatter rather than a separate converter class.",
	},
	{
		UpstreamID: "dates.py:class:rrulewrapper",
		Status:     PublicSurfaceIntentionalOmission,
		GoFiles:    []string{"core/date_tick.go"},
		Note:       "dateutil rrule wrappers are Python-specific and intentionally omitted from the Go API.",
	},
}

// DateSurfaceAuditRows returns the supporting Phase 12 date-surface audit.
func DateSurfaceAuditRows() []DateSurfaceAudit {
	out := make([]DateSurfaceAudit, len(dateSurfaceAuditRows))
	copy(out, dateSurfaceAuditRows)
	for i := range out {
		out[i].GoFiles = append([]string(nil), out[i].GoFiles...)
	}
	return out
}

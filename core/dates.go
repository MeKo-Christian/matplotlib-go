package core

import (
	"errors"
	"time"
)

// Date numbers in matplotlib-go follow Matplotlib's convention: floating-point
// days since an epoch (default 1970-01-01T00:00:00Z). The functions below mirror
// matplotlib.dates.date2num / num2date / set_epoch / get_epoch and are the
// supported way to convert between time.Time and the numeric values plotted on a
// date axis.

// Date2Num converts a time.Time to a date number (days since the current epoch).
func Date2Num(t time.Time) float64 {
	return timeToDateNumber(t)
}

// Num2Date converts a date number (days since the current epoch) back to a
// time.Time in the given location. A nil location is treated as UTC.
func Num2Date(v float64, loc *time.Location) time.Time {
	return dateNumberToTime(v, loc)
}

// GetEpoch returns the current date epoch, resolving the date.epoch rcParam
// if no epoch has been fixed yet. Like Matplotlib's get_epoch, calling it
// locks the epoch: a later SetEpoch fails.
func GetEpoch() time.Time {
	return currentDateEpoch()
}

// SetEpoch sets the reference epoch used for all date<->number conversions.
//
// Like Matplotlib's set_epoch, it must be called before any date conversion or
// date plotting has occurred; otherwise it returns an error and leaves the
// epoch unchanged. It overrides the date.epoch rcParam. Choosing an epoch
// close to the data of interest preserves floating-point precision for
// sub-second resolution far from 1970.
func SetEpoch(t time.Time) error {
	utc := t.UTC()
	if !dateEpochState.CompareAndSwap(nil, &utc) {
		return errors.New("core: SetEpoch must be called before any date conversion or plotting")
	}
	return nil
}

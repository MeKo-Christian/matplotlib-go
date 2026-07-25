package dates_test

import (
	"fmt"
	"time"

	"github.com/cwbudde/matplotlib-go/dates"
)

// Example converts dates to axis values and formats them.
func Example() {
	value := dates.Date2Num(time.Date(2024, time.July, 25, 0, 0, 0, 0, time.UTC))
	fmt.Println(dates.DateFormatter{Layout: "02 Jan 2006", Location: time.UTC}.Format(value))

	// Output:
	// 25 Jul 2024
}

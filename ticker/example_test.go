package ticker_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/ticker"
)

// Example builds evenly spaced ticks and formats them as percentages.
func Example() {
	ticks := (ticker.LinearLocator{NumTicks: 3}).Ticks(0, 1, 3)
	formatter := ticker.PercentFormatter{XMax: 1}
	for i, tick := range ticks {
		fmt.Println(ticker.FormatTick(formatter, tick, i, ticks))
	}

	// Output:
	// 0%
	// 50%
	// 100%
}

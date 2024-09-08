package main

import (
	"fmt"
	"gosignals/signals"
	"time"
)

func main() {
	curr := time.Now()
	totalSleptSecondsSg := signals.MakeSignal(0)
	totalSleptSecondsSg.Listen(func(i int, bs *signals.BaseSignal[int]) {
		fmt.Printf("Total slept seconds: %d\n", i)
	})

	sleep := signals.MakeSignal(0)
	sleep.ListenAsync(func(i int, bs *signals.BaseSignal[int]) {
		time.Sleep(time.Second * time.Duration(i))
		totalSleptSecondsSg.SetFromValue(func(x int) int { return x + i })

		// try this, you will see an unexpected error due to the async operations
		// totalSleptSecondsSg.Set(totalSleptSecondsSg.Get() + i)
	})

	sleep.ListenAsync(func(i int, bs *signals.BaseSignal[int]) {
		time.Sleep(time.Second * time.Duration(i))
		totalSleptSecondsSg.SetFromValue(func(x int) int { return x + i })

		// try this, you will see an unexpected error due to the async operations
		// totalSleptSecondsSg.Set(totalSleptSecondsSg.Get() + i)
	})

	sleep.ListenAsync(func(i int, bs *signals.BaseSignal[int]) {
		time.Sleep(time.Second * time.Duration(i))
		totalSleptSecondsSg.SetFromValue(func(x int) int { return x + i })

		// try this, you will see an unexpected error due to the async operations
		// totalSleptSecondsSg.Set(totalSleptSecondsSg.Get() + i)
	})

	sleep.Set(3)
	// result for ListenAsync impl -> 3 seconds elapsed because each wait 3 seconds but in parallel
	fmt.Printf("Total slept seconds: %v, Time elapsed (including all operations): %v", totalSleptSecondsSg.Get(), time.Since(curr))
}

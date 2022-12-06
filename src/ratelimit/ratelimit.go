package ratelimit

import (
	"math/rand"
	"time"
)

func RunOnly(i int, f func()) {
	randNum := rand.Intn(i) + 1

	if randNum == i {
		f()
	}
}

// Ensure this function is only called once every d duration
func Debounce(d time.Duration, f func()) func() {
	var lastCall time.Time

	return func() {
		if time.Since(lastCall) < d {
			return
		}

		lastCall = time.Now()
		f()
	}
}

func DebounceWithArgs(d time.Duration, f func(interface{}, interface{})) func(interface{}, interface{}) {
	var lastCall time.Time

	return func(arg interface{}, arg2 interface{}) {
		if time.Since(lastCall) < d {
			return
		}

		lastCall = time.Now()
		f(arg, arg2)
	}
}

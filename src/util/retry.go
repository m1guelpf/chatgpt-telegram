package util

import "time"

func Retry(times int, interval time.Duration, fn func() error) error {
	var cnt int
	var err error
	for cnt < times {
		err = fn()
		if err == nil {
			return nil
		}
		cnt++
		time.Sleep(interval)
	}
	return err
}

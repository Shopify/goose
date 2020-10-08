package concurrency

import "time"

func EstimatedWaitTime(limiter Limiter, average time.Duration) time.Duration {
	if average <= 0 {
		return 0
	}

	concurrency := limiter.MaxConcurrency()
	if concurrency == 0 {
		return 0
	}

	return time.Duration(int64(average) / int64(concurrency) * int64(limiter.Waiting()))
}

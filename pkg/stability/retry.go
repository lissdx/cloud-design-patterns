package stability

import (
	"errors"
	"sync"
	"time"
)

const DefaultRetryThreshold = 3

var (
	ErrRetry = errors.New("circuit breaker is open")
)

func defaultRetryExpiryFn(tryCnt int) time.Duration {
	return time.Millisecond * time.Duration((2<<tryCnt)*10)
}

type RetrySettings struct {
	Name           string
	RetryThreshold uint32
	ExpiryFn       func(tryCnt int) time.Duration
}

type Retry struct {
	name           string
	retryThreshold uint32
	expiryFn       func(tryCnt int) time.Duration
	lastAttempt    time.Time
	mutex          sync.Mutex
}


func NewRetry(settings RetrySettings) *Retry {
	retry := new(Retry)

	retry.name = settings.Name

	if retry.retryThreshold = settings.RetryThreshold; retry.retryThreshold <= 0 {
		retry.retryThreshold = DefaultRetryThreshold
	}

	if retry.expiryFn = settings.ExpiryFn; retry.expiryFn == nil {
		retry.expiryFn = defaultRetryExpiryFn
	}

	return retry
}

func (r *Retry) GetProcessorFn(processFn ProcessFn) ProcessFn {
	return func(inObj interface{}) (interface{}, error) {

		for retCnt := 0;; retCnt++ {
			res, err := processFn(inObj)
			if err == nil || uint32(retCnt) >= r.retryThreshold{
				return res, err
			}

			time.Sleep(r.expiryFn(retCnt))
			//select {
			//case <- time.After(r.expiryFn(retCnt)):
			//
			//}
		}
	}
}
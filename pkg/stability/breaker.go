package stability

import (
	"errors"
	"sync"
	"time"
)

const DefaultFailureThreshold = 1

func defaultBreakerExpiryFn(tryCnt int) time.Duration {
	return time.Millisecond * time.Duration((2<<tryCnt)*10)
}

var (
	// ErrOpenState is returned when the CB state is open
	ErrOpenState = errors.New("circuit breaker is open")
)

type BreakerSettings struct {
	Name             string
	FailureThreshold uint32
	ExpiryFn         func(tryCnt int) time.Duration
}

type Breaker struct {
	name                string
	failureThreshold    uint32
	consecutiveFailures uint32
	expiryFn            func(tryCnt int) time.Duration
	lastAttempt         time.Time
	mutex               sync.Mutex
}

func NewBreaker(settings BreakerSettings) *Breaker {
	breaker := new(Breaker)

	breaker.name = settings.Name

	if breaker.failureThreshold = settings.FailureThreshold; breaker.failureThreshold <= 0 {
		breaker.failureThreshold = DefaultFailureThreshold
	}

	if breaker.expiryFn = settings.ExpiryFn; breaker.expiryFn == nil {
		breaker.expiryFn = defaultBreakerExpiryFn
	}

	return breaker
}

func (b *Breaker) GetProcessorFn(processFn ProcessFn) ProcessFn {
	return func(inObj interface{}) (interface{}, error) {

		if err := b.beforeProcess(); err != nil {
			return nil, err
		}

		res, err := processFn(inObj)

		if err := b.afterProcess(err); err != nil {
			return nil, err
		}

		return res, nil
	}
}

func (b *Breaker) afterProcess(err error) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.lastAttempt = time.Now()

	if err != nil {
		b.consecutiveFailures++
	} else {
		b.consecutiveFailures = 0
	}

	return err
}

func (b *Breaker) beforeProcess() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Calculate delta
	d := int(b.consecutiveFailures) - int(b.failureThreshold)

	// If we in Open state
	if d >= 0 {
		now := time.Now()
		shouldRetryAt := b.lastAttempt.Add(b.expiryFn(d))
		// We are in Open state wait
		if now.Before(shouldRetryAt){
			return ErrOpenState
		}
	}

	return nil
}

//func Breaker___(circuit Circuit, failureThreshold uint) Circuit {
//	var consecutiveFailures int = 0
//	var lastAttempt = time.Now()
//	var m sync.RWMutex
//
//	return func(ctx context.Context) (string, error) {
//		m.RLock()                       // Establish a "read lock"
//
//		d := consecutiveFailures - int(failureThreshold)
//
//		if d >= 0 {
//			shouldRetryAt := lastAttempt.Add(time.Second * 2 << d)
//			if !time.Now().After(shouldRetryAt) {
//				m.RUnlock()
//				return "", errors.New("service unreachable")
//			}
//		}
//
//		m.RUnlock()                     // Release read lock
//
//		response, err := circuit(ctx)   // Issue request proper
//
//		m.Lock()                        // Lock around shared resources
//		defer m.Unlock()
//
//		lastAttempt = time.Now()        // Record time of attempt
//
//		if err != nil {                 // Circuit returned an error,
//			consecutiveFailures++       // so we count the failure
//			return response, err        // and return
//		}
//
//		consecutiveFailures = 0         // Reset failures counter
//
//		return response, nil
//	}
//}

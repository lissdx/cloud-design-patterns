package stability

import (
	"fmt"
	"sync"
	"time"
)

// The Throttle pattern is named after a device used to manage the flow of a fluid,
// such as the amount of fuel going into a car engine. Like its namesake mechanism,
// Throttle restricts the number of times that a function can be called during over a period of time.
// For example:
// * A user may only be allowed 10 service requests per second.
// * A client may restrict itself to call a particular function once every 500 milliseconds.
// * An account may only be allowed three failed login attempts in a 24-hour period.

const DefaultMaxTokens = 1
const DefaultRefillTokensCnt = 1
const DefaultRefillInterval = time.Duration(1) * time.Second

type ThrottleSettings struct {
	Name            string
	MaxTokens       uint32
	RefillTokensCnt uint32
	RefillInterval  time.Duration
}

type Throttle struct {
	name            string
	maxTokens       uint32
	tokensInBucket  uint32
	refillTokensCnt uint32
	refillInterval  time.Duration
	once            sync.Once
	mutex           sync.Mutex
}

func (t *Throttle) refill() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.tokensInBucket += t.refillTokensCnt; t.tokensInBucket > t.maxTokens{
		t.tokensInBucket = t.maxTokens
	}
}

func (t *Throttle) isTokenBucketEmpty() bool{
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.tokensInBucket <= 0
}

func (t *Throttle) decreaseTokensInBucket() uint32{
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.tokensInBucket -= 1
	return t.tokensInBucket
}

func NewThrottle(settings ThrottleSettings) *Throttle {
	throttle := new(Throttle)

	throttle.name = settings.Name

	if throttle.maxTokens = settings.MaxTokens; throttle.maxTokens <= 0 {
		throttle.maxTokens = DefaultMaxTokens
	}

	throttle.tokensInBucket = throttle.maxTokens

	if throttle.refillTokensCnt = settings.RefillTokensCnt; throttle.refillTokensCnt <= 0 {
		throttle.refillTokensCnt = DefaultRefillTokensCnt
	}

	if throttle.refillInterval = settings.RefillInterval; throttle.refillInterval <= 0 {
		throttle.refillInterval = DefaultRefillInterval
	}

	return throttle
}

func (t *Throttle) GetProcessorFn(processFn ProcessFn) ProcessFn {

	return func(inObj interface{}) (interface{}, error) {
		t.once.Do(func() {
			ticker := time.NewTicker(t.refillInterval)

			go func() {
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						t.refill()
					}
				}
			}()
		})

		if t.isTokenBucketEmpty() {
			return "", fmt.Errorf("too many calls")
		}

		t.decreaseTokensInBucket()

		return processFn(inObj)
	}
}

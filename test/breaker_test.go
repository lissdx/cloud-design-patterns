package test

import (
	"cloud-design-patterns/pkg/stability"
	"errors"
	"testing"
	"time"
)

var (
	intentionalErr = errors.New("INTENTIONAL FAIL!")
	breakerErr = stability.ErrOpenState
)
//ticker := time.NewTicker(250 * time.Millisecond).C

// failAfter returns a function matching the Circuit type that returns an
// error after its been called more than threshold times.
func failAfter(threshold int) stability.ProcessFn {
	count := 0

	// Service function. Fails after 5 tries.
	return func(inObj interface{}) (interface{}, error) {
		count++

		if count > threshold {
			return nil, intentionalErr
		}

		return inObj, nil
	}
}

// flipFailAfter returns a function matching the Circuit type that returns an
// error / ok and flips the response after flipThreshold
func flipFailAfter(flipThreshold int) stability.ProcessFn {
	count := 0
	isError := true

	// Service function. Fails after 5 tries.
	return func(inObj interface{}) (interface{}, error) {

		if count % flipThreshold == 0 {
			isError = !isError
		}
		count++

		if isError{
			return nil, intentionalErr
		}

		return inObj, nil
	}
}

func TestBreaker(t *testing.T) {

	breakerSettings := stability.BreakerSettings{
		Name: "BreakerSettingsTest", FailureThreshold: 5, ExpiryFn: func(tryCnt int) time.Duration {
			return time.Second * time.Duration(2<<tryCnt)
		},
	}

	breaker := stability.NewBreaker(breakerSettings)
	processorFn := breaker.GetProcessorFn(failAfter(3))
	//fmt.Printf("%v", breaker)

	for i := 0; i < 5; i++ {
		_, err := processorFn(i)
		switch {
		case i < 3 && err != nil:
			t.Error("expected no error; got", err)
		case i > 3 && err == nil:
			t.Error("expected err; got none")
		}
	}

}

func TestBreakerOpenClose(t *testing.T) {

	breakerSettings := stability.BreakerSettings{
		Name: "BreakerSettingsTest", FailureThreshold: 2, ExpiryFn: func(tryCnt int) time.Duration {
			return time.Millisecond * time.Duration(500)
		},
	}


	breaker := stability.NewBreaker(breakerSettings)
	processorFn := breaker.GetProcessorFn(flipFailAfter(3))
	//fmt.Printf("%v", breaker)

	circuitOpen := false
	doesCircuitOpen := false
	doesCircuitReclose := false
	count := 0
	for range time.NewTicker(250 * time.Millisecond).C {
		_, err := processorFn(0)

		if err != nil {
			if err == breakerErr {
				circuitOpen = true
				doesCircuitOpen = true

				t.Log("circuit has opened")
			} else {
				// Does it close again?
				if circuitOpen {
					circuitOpen = false
					doesCircuitReclose = true

					t.Log("circuit has automatically closed")
				}			}
		}

		count++
		if count >= 20 {
			break
		}
	}
	if !doesCircuitOpen {
		t.Error("circuit didn't appear to open")
	}

	if !doesCircuitReclose {
		t.Error("circuit didn't appear to close after time")
	}
}

//func TestThrottleMax10(t *testing.T) {
//
//	calsCnt := func() func(interface{}) (interface{}, error) {
//		cnt := 0
//		return func(_ interface{}) (interface{}, error) {
//			cnt++
//			return cnt, nil
//		}
//	}()
//
//	throttleSettings := stability.ThrottleSettings{
//		Name: "TestThrottleMax1", MaxTokens: 10, RefillTokensCnt: 10, RefillInterval: time.Duration(1) * time.Second,
//	}
//
//	throttle := stability.NewThrottle(throttleSettings)
//	throttleEffector := throttle.GetProcessorFn(calsCnt)
//
//	var resGot int
//	var errCnt int
//
//	for i := 0; i < 100; i++ {
//		res, err := throttleEffector(i)
//		if err == nil {
//			resGot = res.(int)
//		} else {
//			errCnt++
//		}
//
//	}
//
//	if errCnt != 90 {
//		t.Errorf("errCnt = %v, want %v", errCnt, 90)
//	}
//
//	if resGot != 10 {
//		t.Errorf("resGot = %v, want %v", resGot, 10)
//	}
//}
//
//func TestThrottleMax10Refill3Times(t *testing.T) {
//
//	calsCnt := func() func(interface{}) (interface{}, error) {
//		cnt := 0
//		return func(_ interface{}) (interface{}, error) {
//			cnt++
//			return cnt, nil
//		}
//	}()
//
//	throttleSettings := stability.ThrottleSettings{
//		Name: "TestThrottleMax1", MaxTokens: 10, RefillTokensCnt: 10, RefillInterval: time.Duration(300) * time.Millisecond,
//	}
//
//	throttle := stability.NewThrottle(throttleSettings)
//	throttleEffector := throttle.GetProcessorFn(calsCnt)
//
//	var resGot int
//	var errCnt int
//
//	for a := 0; a < 3; a++ {
//		for i := 0; i < 100; i++ {
//			res, err := throttleEffector(i)
//			if err == nil {
//				resGot = res.(int)
//			} else {
//				errCnt++
//			}
//		}
//		//fmt.Printf("resGot: %d, errCnt: %d\n", resGot, errCnt)
//		time.Sleep(time.Duration(1) * time.Second)
//	}
//
//	if errCnt != 270 {
//		t.Errorf("errCnt = %v, want %v", errCnt, 270)
//	}
//
//	if resGot != 30 {
//		t.Errorf("resGot = %v, want %v", resGot, 30)
//	}
//}
//
//func TestThrottleQuickRefill(t *testing.T) {
//
//	calsCnt := func() func(interface{}) (interface{}, error) {
//		cnt := 0
//		return func(_ interface{}) (interface{}, error) {
//			cnt++
//			return cnt, nil
//		}
//	}()
//
//	throttleSettings := stability.ThrottleSettings{
//		Name: "TestThrottleMax1", MaxTokens: 10, RefillTokensCnt: 10, RefillInterval: time.Duration(10) * time.Millisecond,
//	}
//
//	throttle := stability.NewThrottle(throttleSettings)
//	throttleEffector := throttle.GetProcessorFn(calsCnt)
//
//	var resGot int
//	var errCnt int
//
//	for i := 0; i < 100; i++ {
//		res, err := throttleEffector(i)
//		if err == nil {
//			resGot = res.(int)
//		} else {
//			errCnt++
//		}
//		time.Sleep(time.Duration(1) * time.Millisecond)
//	}
//	fmt.Printf("resGot: %d, errCnt: %d\n", resGot, errCnt)
//
//
//	if errCnt != 0 {
//		t.Errorf("errCnt = %v, want %v", errCnt, 0)
//	}
//
//	if resGot != 100 {
//		t.Errorf("resGot = %v, want %v", resGot, 100)
//	}
//}
//
//func TestThrottleFrequency5Seconds(t *testing.T) {
//
//	calsCnt := func() func(interface{}) (interface{}, error) {
//		cnt := 0
//		return func(_ interface{}) (interface{}, error) {
//			cnt++
//			return cnt, nil
//		}
//	}()
//
//	throttleSettings := stability.ThrottleSettings{
//		Name: "TestThrottleMax1", MaxTokens: 1, RefillTokensCnt: 1, RefillInterval: time.Duration(1) * time.Second,
//	}
//
//	throttle := stability.NewThrottle(throttleSettings)
//	throttleEffector := throttle.GetProcessorFn(calsCnt)
//
//	var resGot int
//	var errCnt int
//	ticker := time.NewTicker(250 * time.Millisecond).C
//	tickCounts := 0
//
//	for range ticker {
//		tickCounts++
//		res, err := throttleEffector(0)
//		if err == nil {
//			resGot = res.(int)
//		} else {
//			errCnt++
//		}
//
//		if tickCounts >= 20 {
//			break
//		}
//	}
//
//	fmt.Printf("resGot: %d, errCnt: %d\n", resGot, errCnt)
//
//	if errCnt != 15 {
//		t.Errorf("errCnt = %v, want %v", errCnt, 15)
//	}
//
//	if resGot != 5 {
//		t.Errorf("resGot = %v, want %v", resGot, 5)
//	}
//}
//
//func TestThrottleRefill(t *testing.T) {
//
//	calsCnt := func() func(interface{}) (interface{}, error) {
//		cnt := 0
//		return func(_ interface{}) (interface{}, error) {
//			cnt++
//			return cnt, nil
//		}
//	}()
//
//	throttleSettings := stability.ThrottleSettings{
//		Name: "TestThrottleMax1", MaxTokens: 5, RefillTokensCnt: 1, RefillInterval: time.Duration(1) * time.Second,
//	}
//
//	throttle := stability.NewThrottle(throttleSettings)
//	throttleEffector := throttle.GetProcessorFn(calsCnt)
//
//	var resGot int
//	var errCnt int
//	for i := 0; i < 10; i++ {
//		res, err := throttleEffector(i)
//		if err == nil {
//			resGot = res.(int)
//		} else {
//			errCnt++
//		}
//	}
//	time.Sleep(time.Duration(3500) * time.Millisecond)
//
//	for i := 0; i < 10; i++ {
//		res, err := throttleEffector(i)
//		if err == nil {
//			resGot = res.(int)
//		} else {
//			errCnt++
//		}
//	}
//
//
//	fmt.Printf("resGot: %d, errCnt: %d\n", resGot, errCnt)
//
//	if errCnt != 12 {
//		t.Errorf("errCnt = %v, want %v", errCnt, 12)
//	}
//
//	if resGot != 8 {
//		t.Errorf("resGot = %v, want %v", resGot, 8)
//	}
//}

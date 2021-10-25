package test

import (
	"cloud-design-patterns/pkg/stability"
	"fmt"
	"testing"
	"time"
)

func TestThrottleMax1(t *testing.T) {

	calsCnt := func() func(interface{}) (interface{}, error) {
		cnt := 0
		return func(_ interface{}) (interface{}, error) {
			cnt++
			return cnt, nil
		}
	}()

	throttleSettings := stability.ThrottleSettings{
		Name: "TestThrottleMax1", MaxTokens: 1, RefillTokensCnt: 1, RefillInterval: time.Duration(1) * time.Second,
	}

	throttle := stability.NewThrottle(throttleSettings)
	throttleEffector := throttle.GetProcessorFn(calsCnt)

	var resGot int
	var errCnt int

	for i := 0; i < 100; i++ {
		res, err := throttleEffector(i)
		if err == nil {
			resGot = res.(int)
		} else {
			errCnt++
		}

	}

	if errCnt != 99 {
		t.Errorf("errCnt = %v, want %v", errCnt, 99)
	}

	if resGot != 1 {
		t.Errorf("resGot = %v, want %v", resGot, 1)
	}
}

func TestThrottleMax10(t *testing.T) {

	calsCnt := func() func(interface{}) (interface{}, error) {
		cnt := 0
		return func(_ interface{}) (interface{}, error) {
			cnt++
			return cnt, nil
		}
	}()

	throttleSettings := stability.ThrottleSettings{
		Name: "TestThrottleMax1", MaxTokens: 10, RefillTokensCnt: 10, RefillInterval: time.Duration(1) * time.Second,
	}

	throttle := stability.NewThrottle(throttleSettings)
	throttleEffector := throttle.GetProcessorFn(calsCnt)

	var resGot int
	var errCnt int

	for i := 0; i < 100; i++ {
		res, err := throttleEffector(i)
		if err == nil {
			resGot = res.(int)
		} else {
			errCnt++
		}

	}

	if errCnt != 90 {
		t.Errorf("errCnt = %v, want %v", errCnt, 90)
	}

	if resGot != 10 {
		t.Errorf("resGot = %v, want %v", resGot, 10)
	}
}

func TestThrottleMax10Refill3Times(t *testing.T) {

	calsCnt := func() func(interface{}) (interface{}, error) {
		cnt := 0
		return func(_ interface{}) (interface{}, error) {
			cnt++
			return cnt, nil
		}
	}()

	throttleSettings := stability.ThrottleSettings{
		Name: "TestThrottleMax1", MaxTokens: 10, RefillTokensCnt: 10, RefillInterval: time.Duration(300) * time.Millisecond,
	}

	throttle := stability.NewThrottle(throttleSettings)
	throttleEffector := throttle.GetProcessorFn(calsCnt)

	var resGot int
	var errCnt int

	for a := 0; a < 3; a++ {
		for i := 0; i < 100; i++ {
			res, err := throttleEffector(i)
			if err == nil {
				resGot = res.(int)
			} else {
				errCnt++
			}
		}
		//fmt.Printf("resGot: %d, errCnt: %d\n", resGot, errCnt)
		time.Sleep(time.Duration(1) * time.Second)
	}

	if errCnt != 270 {
		t.Errorf("errCnt = %v, want %v", errCnt, 270)
	}

	if resGot != 30 {
		t.Errorf("resGot = %v, want %v", resGot, 30)
	}
}

func TestThrottleQuickRefill(t *testing.T) {

	calsCnt := func() func(interface{}) (interface{}, error) {
		cnt := 0
		return func(_ interface{}) (interface{}, error) {
			cnt++
			return cnt, nil
		}
	}()

	throttleSettings := stability.ThrottleSettings{
		Name: "TestThrottleMax1", MaxTokens: 10, RefillTokensCnt: 10, RefillInterval: time.Duration(10) * time.Millisecond,
	}

	throttle := stability.NewThrottle(throttleSettings)
	throttleEffector := throttle.GetProcessorFn(calsCnt)

	var resGot int
	var errCnt int

	for i := 0; i < 100; i++ {
		res, err := throttleEffector(i)
		if err == nil {
			resGot = res.(int)
		} else {
			errCnt++
		}
		time.Sleep(time.Duration(1) * time.Millisecond)
	}
	fmt.Printf("resGot: %d, errCnt: %d\n", resGot, errCnt)


	if errCnt != 0 {
		t.Errorf("errCnt = %v, want %v", errCnt, 0)
	}

	if resGot != 100 {
		t.Errorf("resGot = %v, want %v", resGot, 100)
	}
}

func TestThrottleFrequency5Seconds(t *testing.T) {

	calsCnt := func() func(interface{}) (interface{}, error) {
		cnt := 0
		return func(_ interface{}) (interface{}, error) {
			cnt++
			return cnt, nil
		}
	}()

	throttleSettings := stability.ThrottleSettings{
		Name: "TestThrottleMax1", MaxTokens: 1, RefillTokensCnt: 1, RefillInterval: time.Duration(1) * time.Second,
	}

	throttle := stability.NewThrottle(throttleSettings)
	throttleEffector := throttle.GetProcessorFn(calsCnt)

	var resGot int
	var errCnt int
	ticker := time.NewTicker(250 * time.Millisecond).C
	tickCounts := 0

	for range ticker {
		tickCounts++
		res, err := throttleEffector(0)
		if err == nil {
			resGot = res.(int)
		} else {
			errCnt++
		}

		if tickCounts >= 20 {
			break
		}
	}

	fmt.Printf("resGot: %d, errCnt: %d\n", resGot, errCnt)

	if errCnt != 15 {
		t.Errorf("errCnt = %v, want %v", errCnt, 15)
	}

	if resGot != 5 {
		t.Errorf("resGot = %v, want %v", resGot, 5)
	}
}

func TestThrottleRefill(t *testing.T) {

	calsCnt := func() func(interface{}) (interface{}, error) {
		cnt := 0
		return func(_ interface{}) (interface{}, error) {
			cnt++
			return cnt, nil
		}
	}()

	throttleSettings := stability.ThrottleSettings{
		Name: "TestThrottleMax1", MaxTokens: 5, RefillTokensCnt: 1, RefillInterval: time.Duration(1) * time.Second,
	}

	throttle := stability.NewThrottle(throttleSettings)
	throttleEffector := throttle.GetProcessorFn(calsCnt)

	var resGot int
	var errCnt int
	for i := 0; i < 10; i++ {
		res, err := throttleEffector(i)
		if err == nil {
			resGot = res.(int)
		} else {
			errCnt++
		}
	}
	time.Sleep(time.Duration(3500) * time.Millisecond)

	for i := 0; i < 10; i++ {
		res, err := throttleEffector(i)
		if err == nil {
			resGot = res.(int)
		} else {
			errCnt++
		}
	}


	fmt.Printf("resGot: %d, errCnt: %d\n", resGot, errCnt)

	if errCnt != 12 {
		t.Errorf("errCnt = %v, want %v", errCnt, 12)
	}

	if resGot != 8 {
		t.Errorf("resGot = %v, want %v", resGot, 8)
	}
}

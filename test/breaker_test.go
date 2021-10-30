package test

import (
	"cloud-design-patterns/pkg/stability"
	"errors"
	"fmt"
	"github.com/lissdx/yapgo/pkg/pipeline"
	"sync"
	"testing"
	"time"
)

var (
	intentionalErr = errors.New("INTENTIONAL FAIL!")
	breakerErr = stability.ErrOpenState
)

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
	var mu sync.Mutex
	// Service function. Fails after 5 tries.
	return func(inObj interface{}) (interface{}, error) {

		mu.Lock()
		if count % flipThreshold == 0 {
			isError = !isError
		}
		count++
		mu.Unlock()

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

// errHandlerChan channel which has
// personal errorHandler method
type errHandlerChan pipeline.BidirectionalStream
func (e errHandlerChan) getErrorHandlerFn() pipeline.ErrorProcessFn{
	return func(err error) {
		e <- err
	}
}

// TestCircuitBreakerDataRace tests for data races.
func TestCircuitBreakerDataRace(t *testing.T) {
	eh := make(errHandlerChan) // Create error handler
	defer close(eh)
	doneCh := make(chan interface{}) // doneChannel (control channel)
	flipThreshold := 6
	fanSize := 3
	failureThreshold := 3
	breakerSettings := stability.BreakerSettings{ // our circleBreaker settings
		Name:             "TestCircuitBreakerDataRace",
		FailureThreshold: uint32(failureThreshold),
		ExpiryFn: func(tryCnt int) time.Duration {
			return time.Millisecond * time.Duration(500)
		},
	}
	// function that works under circle breaker
	processorFn := stability.NewBreaker(breakerSettings).GetProcessorFn(flipFailAfter(flipThreshold))
	vals := []interface{}{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24}
	// Create pipeline
	pl := pipeline.New()
	// Create funOut stage in pipeline
	pl.AddStageWithFanOut(pipeline.ProcessFn(processorFn),  eh.getErrorHandlerFn(),uint64(fanSize))
	// Run pipeline and send output (result) to channel
	outChan := pl.Run(doneCh, pipeline.Generator(doneCh, vals...))


	funcErrCnt := 0
	breakerErrCnt := 0
	resCnt := 0
	for i := 0; i < len(vals); i++ {
		select {
		case v := <- outChan:
			t.Log(fmt.Sprintf("res: %v", v))
			resCnt++
		case e := <- eh:
			t.Log(fmt.Sprintf("err: %v", e))
			switch e == breakerErr {
			case true:
				breakerErrCnt++
				time.Sleep(time.Second * 1)
			default:
				funcErrCnt++
			}
		}
	}
	//time.Sleep(time.Second * 2)
	wantRes := 12
	wantBreakerErr := 4
	wantFuncErr := 8
	t.Logf("funcErrCnt: %d", funcErrCnt)
	t.Logf("breakerErrCnt: %d", breakerErrCnt)
	t.Logf("resCnt: %d", resCnt)
	if funcErrCnt != wantFuncErr {
		t.Errorf("funcErrCnt error. want %d got %d", wantFuncErr, funcErrCnt )
	}
	if resCnt != wantRes {
		t.Errorf("resCnt error. want %d got %d", wantRes, resCnt )
	}
	if breakerErrCnt != wantBreakerErr {
		t.Errorf("breakerErrCnt error. want %d got %d", wantBreakerErr, breakerErrCnt)
	}
}


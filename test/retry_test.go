package test

import (
	"cloud-design-patterns/pkg/stability"
	"errors"
	"fmt"
	"testing"
)

var (
	retryErr = errors.New("INTENTIONAL FAIL!")
)

func EmulateHttHttpsConnection() stability.ProcessFn {
	protocol := []interface{}{"http", "https", "ftp"}
	curProtocolInd := 0
	callCnt := 1
	return func(inObj interface{}) (interface{}, error) {
		url := inObj.(string)
		if callCnt >= 3 {
			return fmt.Sprintf("%s://%s", protocol[curProtocolInd], url), nil
		}
		callCnt++
		curProtocolInd++
		curProtocolInd %= 3

		return inObj, retryErr
	}

}
func TestRetry(t *testing.T) {

	vals := []interface{}{"url1.com", "url2.com", "url3.com", "url4.com", "url5.com"}
	retrySettings := stability.RetrySettings{
		Name: "BreakerSettingsTest", RetryThreshold: 5, ExpiryFn: nil,
	}

	retry := stability.NewRetry(retrySettings)
	processorFn := retry.GetProcessorFn(EmulateHttHttpsConnection())
	//fmt.Printf("%v", breaker)

	errCnt := 0
	resCnt := 0
	for _, u := range vals {
		res, err := processorFn(u)
		switch {
		case err != nil:
			t.Log(err)
			errCnt++
		default:
			t.Logf("res: %v", res)
			resCnt++
		}
	}
	if resCnt <= 0 {
		t.Error("retry test must return result")
	}
}

//
//func TestBreakerOpenClose(t *testing.T) {
//
//	breakerSettings := stability.BreakerSettings{
//		Name: "BreakerSettingsTest", FailureThreshold: 2, ExpiryFn: func(tryCnt int) time.Duration {
//			return time.Millisecond * time.Duration(500)
//		},
//	}
//
//
//	breaker := stability.NewBreaker(breakerSettings)
//	processorFn := breaker.GetProcessorFn(flipFailAfter(3))
//	//fmt.Printf("%v", breaker)
//
//	circuitOpen := false
//	doesCircuitOpen := false
//	doesCircuitReclose := false
//	count := 0
//	for range time.NewTicker(250 * time.Millisecond).C {
//		_, err := processorFn(0)
//
//		if err != nil {
//			if err == breakerErr {
//				circuitOpen = true
//				doesCircuitOpen = true
//
//				t.Log("circuit has opened")
//			} else {
//				// Does it close again?
//				if circuitOpen {
//					circuitOpen = false
//					doesCircuitReclose = true
//
//					t.Log("circuit has automatically closed")
//				}			}
//		}
//
//		count++
//		if count >= 20 {
//			break
//		}
//	}
//	if !doesCircuitOpen {
//		t.Error("circuit didn't appear to open")
//	}
//
//	if !doesCircuitReclose {
//		t.Error("circuit didn't appear to close after time")
//	}
//}
//
//// errHandlerChan channel which has
//// personal errorHandler method
//type errHandlerChan pipeline.BidirectionalStream
//func (e errHandlerChan) getErrorHandlerFn() pipeline.ErrorProcessFn{
//	return func(err error) {
//		e <- err
//	}
//}
//
//// TestCircuitBreakerDataRace tests for data races.
//func TestCircuitBreakerDataRace(t *testing.T) {
//	eh := make(errHandlerChan) // Create error handler
//	defer close(eh)
//	doneCh := make(chan interface{}) // doneChannel (control channel)
//	flipThreshold := 6
//	fanSize := 3
//	failureThreshold := 3
//	breakerSettings := stability.BreakerSettings{ // our circleBreaker settings
//		Name:             "TestCircuitBreakerDataRace",
//		FailureThreshold: uint32(failureThreshold),
//		ExpiryFn: func(tryCnt int) time.Duration {
//			return time.Millisecond * time.Duration(500)
//		},
//	}
//	// function that works under circle breaker
//	processorFn := stability.NewBreaker(breakerSettings).GetProcessorFn(flipFailAfter(flipThreshold))
//	vals := []interface{}{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24}
//	// Create pipeline
//	pl := pipeline.New()
//	// Create funOut stage in pipeline
//	pl.AddStageWithFanOut(pipeline.ProcessFn(processorFn),  eh.getErrorHandlerFn(),uint64(fanSize))
//	// Run pipeline and send output (result) to channel
//	outChan := pl.Run(doneCh, pipeline.Generator(doneCh, vals...))
//
//
//	funcErrCnt := 0
//	breakerErrCnt := 0
//	resCnt := 0
//	for i := 0; i < len(vals); i++ {
//		select {
//		case v := <- outChan:
//			t.Log(fmt.Sprintf("res: %v", v))
//			resCnt++
//		case e := <- eh:
//			t.Log(fmt.Sprintf("err: %v", e))
//			switch e == breakerErr {
//			case true:
//				breakerErrCnt++
//				time.Sleep(time.Second * 1)
//			default:
//				funcErrCnt++
//			}
//		}
//	}
//	//time.Sleep(time.Second * 2)
//	wantRes := 12
//	wantBreakerErr := 4
//	wantFuncErr := 8
//	t.Logf("funcErrCnt: %d", funcErrCnt)
//	t.Logf("breakerErrCnt: %d", breakerErrCnt)
//	t.Logf("resCnt: %d", resCnt)
//	if funcErrCnt != wantFuncErr {
//		t.Errorf("funcErrCnt error. want %d got %d", wantFuncErr, funcErrCnt )
//	}
//	if resCnt != wantRes {
//		t.Errorf("resCnt error. want %d got %d", wantRes, resCnt )
//	}
//	if breakerErrCnt != wantBreakerErr {
//		t.Errorf("breakerErrCnt error. want %d got %d", wantBreakerErr, breakerErrCnt)
//	}
//}
//

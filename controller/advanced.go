package controller

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/redhat-appstudio-qe/concurency-controller/utils"
)

var counter int64




func (l LoadController) ExecuteSpike(testFunction utils.RunnerFunction) {
    var (
        wg sync.WaitGroup
        start = time.Now()
        rps = 1
        maxRPS = l.maxRPS
        errorThreshold = l.errorThreshold
        errorCount = 0
        totalRequests = 0
        totalLatency time.Duration
        maxLatency time.Duration
        minLatency time.Duration = time.Hour
    )
    // specify the timeout duration here
    timeout := l.timeout

    timer := time.NewTimer(timeout)
    defer timer.Stop()
    for {
        select {
        case <-timer.C:
            fmt.Printf("Test timed out after %v\n", timeout)
            fmt.Printf("Total requests made: %d\n", totalRequests)
            fmt.Printf("Requests per second: %d\n", rps)
            fmt.Printf("Error count: %d\n", errorCount)
            fmt.Printf("Average Latency: %v\n", totalLatency / time.Duration(totalRequests))
            fmt.Printf("Max Latency: %v\n", maxLatency)
            fmt.Printf("Min Latency: %v\n", minLatency)
			l.Results.Latency = totalLatency / time.Duration(totalRequests)
			l.Results.RPS = rps
			l.Results.TotalRequests = totalRequests
			l.Results.TotalErrors = errorCount
            return
        default:
            wg.Add(1)
			
            go func() {
                defer wg.Done()
                start := time.Now() // added this line
                if err := utils.PointerRunnerWrap(testFunction, &counter); err != nil {
                    errorCount++
                }
                totalRequests++
                latency := time.Since(start) // added this line
                totalLatency += latency // added this line
                if latency > maxLatency { // added this line
                    maxLatency = latency // added this line
                }
                if latency < minLatency { // added this line
                    minLatency = latency // added this line
                }
            }()
			//l.Results.Latency = totalLatency / time.Duration(totalRequests)
        }

        if totalRequests%rps == 0 {
            time.Sleep(time.Second) // added this line to wait 1 sec
            duration := time.Since(start) // added this line
			counter = 0
            fmt.Printf("RPS: %d\n", int(float64(rps)/duration.Seconds())) // added this line
			l.Results.RPS = rps
			l.Results.TotalRequests = totalRequests
			l.Results.TotalErrors = errorCount
			if totalLatency != 0 {
				l.Results.Latency = totalLatency / time.Duration(totalRequests)
			}
            errorRate := float64(errorCount) / float64(totalRequests)
            if errorRate > errorThreshold {
                rps = int(math.Max(float64(rps/2), 1)) // added this line
                errorCount = 0
				//break // added this line
            }
            if rps < maxRPS {
                rps = int(math.Min(float64(rps*2), float64(maxRPS)))
            }
            start = time.Now()
        }
        wg.Wait()
	}
}


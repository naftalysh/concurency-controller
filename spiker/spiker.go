package spiker

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	typesdef "github.com/redhat-appstudio-qe/concurency-controller/typesDef"
)

type Spiker struct {
	maxRPS int
	timeout time.Duration
	errorThresholdRate float64
	Results *typesdef.Results
}

func NewSpiker(maxRPS int, timeout time.Duration, errorThresholdRate float64, results *typesdef.Results) *Spiker {
	return &Spiker{
		maxRPS: maxRPS,
		timeout: timeout,
		errorThresholdRate: errorThresholdRate,
		Results: results,
	}
}

func loadTest(runner typesdef.RunnerFunction, rps int, wg *sync.WaitGroup, totalRequests *int, totalFinished *int, totalErrors *int, avgResponseTime *time.Duration) {
	start := time.Now()
	err := runner()
	elapsed := time.Since(start)
	*totalFinished++
	*avgResponseTime = time.Duration(((*avgResponseTime*time.Duration(*totalFinished-1) + elapsed) / time.Duration(*totalFinished)))
	if err != nil {
		*totalErrors++
	}
	wg.Done()
}

func (s Spiker) GenerateSpike(runner typesdef.RunnerFunction) {
	var wg sync.WaitGroup
	var totalRequests int
	var totalFinished int
	var totalErrors int
	var avgResponseTime time.Duration
	var avgRPS float64
	rps := 1
	timeout := time.After(s.timeout)
	log.Println("Starting With RPS: ", rps)
	start := time.Now()
	for {
		select {
		case <-timeout:
			goto finish
		default:
			for i := 0; i < rps; i++ {
				wg.Add(1)
				totalRequests++
				log.Println("Trigger")
				go loadTest(runner, rps, &wg, &totalRequests, &totalFinished, &s.Results.TotalErrors, &s.Results.Latency)
				log.Println("Finish")
				s.Results.TotalRequests = totalRequests
				
			}
			time.Sleep(time.Second)
			errorRate := float64(totalErrors) / float64(totalFinished)
			avgRPS = float64(totalRequests) / time.Since(start).Seconds()
			s.Results.RPS = rps
			if totalRequests%10 == 0 && errorRate < s.errorThresholdRate && rps < s.maxRPS {
				rps = int(math.Min(float64(rps*2), float64(s.maxRPS)))
				log.Println("RPS increased to:", rps)
			}
			if errorRate >= s.errorThresholdRate {
				rps = int(math.Max(float64(rps/2), 1))
				log.Println("RPS decreased to:", rps)
			}
		}
	}
finish:
	wg.Wait()
	fmt.Println("Total Requests:", totalRequests)
	fmt.Println("Total Requests Finished:", totalFinished)
	fmt.Println("Total Errors:", totalErrors)
	fmt.Println("Average Response Time:", avgResponseTime)
	fmt.Println("Average RPS:", avgRPS)

}
package consumer

import (
	"log"
	"sync"
	"time"

	"github.com/redhat-appstudio-qe/concurency-controller/typesDef"
)

type Consumer struct {
	Active *chan int
	Results *typesdef.Results
	Wg sync.WaitGroup
}


var (
	AverageTimeForBatch  time.Duration
	TotalReq int = 0
	TotalFailedReq int = 0
)

// NewConsumer creates a Consumer
func NewConsumer(active *chan int, results *typesdef.Results) *Consumer {
	TotalReq, TotalFailedReq = 0, 0
	return &Consumer{Active: active, Results: results}
}

// consume reads the msgs channel
func (c *Consumer) Consume(RPS int, runner typesdef.RunnerFunction , Batches int,  monitoringURL string, sendMetrics bool) {
	c.Wg = sync.WaitGroup{}
	log.Println("consume: Started")
	for {
		c.CommonConsume(RPS, runner, Batches, monitoringURL, sendMetrics)
	}
}


func (c *Consumer) CommonConsume(RPS int, runner typesdef.RunnerFunction , Batches int,  monitoringURL string, sendMetrics bool) {
	active_thread := <-*c.Active
	log.Println("consume: Received:", active_thread)
	startTime := time.Now()
	for j := 0; j<RPS; j++ {
		c.Wg.Add(1)
		go func(id int){
			TotalReq += 1
			err := typesdef.RunnerWrap(runner, id, active_thread)
			if err != nil {
				TotalFailedReq += 1
			}
		}(j)
		c.Wg.Done()
		time.Sleep(time.Microsecond * 25)
	}
	Endtime := time.Since(startTime)
	AverageTimeForBatch += Endtime
	c.Wg.Wait()
	c.Results.TotalRequests = TotalReq
	c.Results.TotalErrors = TotalFailedReq
	c.Results.RPS = RPS
	if AverageTimeForBatch != 0{
		c.Results.Latency = AverageTimeForBatch / time.Duration(TotalReq)
	}
	log.Println(c.Results)
	metricsPrinter(active_thread,Batches,Endtime,AverageTimeForBatch,TotalReq,TotalFailedReq,monitoringURL, sendMetrics)
}

func metricsPrinter(active_thread int, 
	Batches int, Endtime time.Duration, 
	AverageTimeForBatch time.Duration, 
	TotalReq int, 
	TotalFailedReq int, monitoringURL string ,sendMetrics bool){
	log.Println("Total Time taken for this batch: ", Endtime)
	log.Println("Latency: ", AverageTimeForBatch / time.Duration(TotalReq))
	log.Println("Requests Counter: ", TotalReq)
	log.Println("Successful Requests Counter: ", TotalReq - TotalFailedReq)
	log.Println("Failed Requests Counter: ", TotalFailedReq)
	
}
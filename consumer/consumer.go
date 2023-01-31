package consumer

import (
	"log"
	"sync"
	"time"

	"github.com/redhat-appstudio-qe/concurency-controller/utils"
)

type Consumer struct {
	active *chan int
	wg sync.WaitGroup
}

var (
	AverageTimeForBatch time.Duration
	TotalReq int = 0
	TotalPassedReq int = 0
	TotalFailedReq int = 0
)

// NewConsumer creates a Consumer
func NewConsumer(active *chan int) *Consumer {
	TotalReq, TotalFailedReq, TotalPassedReq = 0 , 0, 0
	return &Consumer{active: active}
}

// consume reads the msgs channel
func (c *Consumer) Consume(RPS int, runner utils.RunnerFunction, Batches int,  monitoringURL string, sendMetrics bool) {
	c.wg = sync.WaitGroup{}
	log.Println("consume: Started")
	for {
		t := c.CommonConsume(RPS, runner, Batches, monitoringURL, sendMetrics)
		if sendMetrics{
			utils.SendTotal(float64(t), utils.GetPath(monitoringURL,"updateTotal"))
		}
	}
}


func (c *Consumer) CommonConsume(RPS int, runner utils.RunnerFunction , Batches int,  monitoringURL string, sendMetrics bool) int {
	active_thread := <-*c.active
	log.Println("consume: Received:", active_thread)
	startTime := time.Now()
	for j := 0; j<RPS; j++ {
		c.wg.Add(1)
		go func(id int){
			TotalReq += 1
			err := utils.RunnerWrap(runner, id, active_thread)
			if err != nil {
				TotalFailedReq += 1
			} else{
				TotalPassedReq += 1
			}
		}(j)
		c.wg.Done()
		time.Sleep(time.Microsecond * 25)
	}
	Endtime := time.Since(startTime)
	AverageTimeForBatch += Endtime
	c.wg.Wait()
	metricsPrinter(active_thread,Batches,Endtime,AverageTimeForBatch,TotalReq,TotalPassedReq,TotalFailedReq,monitoringURL, sendMetrics)
	return TotalReq
}

func metricsPrinter(active_thread int, 
	Batches int, Endtime time.Duration, 
	AverageTimeForBatch time.Duration, 
	TotalReq int, TotalPassedReq int, 
	TotalFailedReq int, monitoringURL string ,sendMetrics bool){
	log.Println("Total Time taken for this batch: ", Endtime)
	log.Println("Avg time taken by this batch: ", AverageTimeForBatch)
	log.Println("Requests Counter: ", TotalReq)
	log.Println("Successful Requests Counter: ", TotalPassedReq)
	log.Println("Failed Requests Counter: ", TotalFailedReq)
	if sendMetrics{
		utils.SendMetrics(float64(TotalReq), float64(TotalPassedReq), float64(TotalFailedReq), utils.GetPath(monitoringURL,"addBatchWise"))
		utils.SendTime(float64(AverageTimeForBatch) / float64(time.Millisecond),  utils.GetPath(monitoringURL,"updateAvgTime"))
	}
}
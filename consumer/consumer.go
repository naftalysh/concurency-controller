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

type SpikeConsumer struct {
	active *chan int
	done *chan bool
	RPS *int
	wg sync.WaitGroup
}

type ConsumerFunction func(int, int)(error)

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

// NewConsumer creates a Consumer
func NewConsumer2(active *chan int) *SpikeConsumer{
	TotalReq, TotalFailedReq, TotalPassedReq = 0 , 0, 0
	return &SpikeConsumer{active: active}
}



// consume reads the msgs channel
func (c *SpikeConsumer) Consume2(runner ConsumerFunction, Batches int,  monitoringURL string, sendMetrics bool) {
	c.wg = sync.WaitGroup{}
	log.Println("consume: Started")
	for {
		t := c.CommonConsume2(2, runner, Batches, monitoringURL, sendMetrics)
		if sendMetrics{
			utils.SendTotal(float64(t), utils.GetPath(monitoringURL,"updateTotal"))
		}
	}
}

// consume reads the msgs channel
func (c *Consumer) Consume(RPS int, runner ConsumerFunction, Batches int,  monitoringURL string, sendMetrics bool) {
	c.wg = sync.WaitGroup{}
	log.Println("consume: Started")
	for {
		t := c.CommonConsume(RPS, runner, Batches, monitoringURL, sendMetrics)
		if sendMetrics{
			utils.SendTotal(float64(t), utils.GetPath(monitoringURL,"updateTotal"))
		}
	}
}

func RunnerWrap(TotalFailedReq *int , TotalPassedReq *int , runner ConsumerFunction, RPS int , i ...int) (int) {
	err := runner(i[0], i[1])
	if err != nil {
		*TotalFailedReq += 1
		return RPS - 2
	}
	*TotalPassedReq += 1
	return RPS + 2
}


func (c *SpikeConsumer) CommonConsume2(minRPS int, runner ConsumerFunction, Batches int,  monitoringURL string, sendMetrics bool) int {
	active_thread := <-*c.active
    
	active_RPS := minRPS

	*c.RPS = active_RPS
	
	log.Println("consume: Received:", active_thread, "RPS", active_RPS)
	startTime := time.Now()
	for j := 0; j<active_RPS; j++ {
		c.wg.Add(1)
		go func(id int){
			
			TotalReq += 1
			active_RPS = RunnerWrap(&TotalFailedReq, &TotalPassedReq, runner, active_RPS, id, active_thread)
		
		}(j)
		c.wg.Done()
		time.Sleep(time.Microsecond * 15)
	}
	Endtime := time.Since(startTime)
	AverageTimeForBatch += Endtime
	c.wg.Wait()
	//metricsPrinter(active_thread,Batches,Endtime,AverageTimeForBatch,TotalReq,TotalPassedReq,TotalFailedReq,monitoringURL, sendMetrics)
	return TotalReq
}

func (c *Consumer) CommonConsume(RPS int, runner ConsumerFunction, Batches int,  monitoringURL string, sendMetrics bool) int {
	active_thread := <-*c.active
	log.Println("consume: Received:", active_thread)
	startTime := time.Now()
	for j := 0; j<RPS; j++ {
		c.wg.Add(1)
		go func(id int){
			TotalReq += 1
			err := runner(id, active_thread)
			if err != nil {
				TotalFailedReq += 1
			} else{
				TotalPassedReq += 1
			}
		}(j)
		c.wg.Done()
		time.Sleep(time.Microsecond * 15)
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
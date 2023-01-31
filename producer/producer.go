package producer

import (
	"time"
	"log"
	"github.com/redhat-appstudio-qe/concurency-controller/utils"
)

// Producer definition
type Producer struct {
	active *chan int
	done *chan bool
}

var (
	TotalTime time.Duration
)

// NewProducer creates a Producer
func NewProducer(active *chan int, done *chan bool) *Producer {
	return &Producer{active: active, done: done}
}

func (p *Producer) ProduceInfinite(timeout time.Duration, monitoringURL string, sendMetrics bool){
	log.Println("produce: Started")
	startTime := time.Now()
	var i int =0
	for  start := time.Now(); time.Since(start) < timeout; {
		log.Println("produce: Sending ", i)
		*p.active <- i
		time.Sleep(time.Second * 1)
		i++
	}
	TotalTime = time.Since(startTime)
	if sendMetrics{
		utils.SendTime(float64(TotalTime) / float64(time.Second),  utils.GetPath(monitoringURL, "updateTime"))
	}
	log.Println("produce: Done/ time taken: ", TotalTime)
	*p.done <- true // signal when done
}

func (p *Producer) Produce(max int,  monitoringURL string, sendMetrics bool) {
	log.Println("produce: Started")
	startTime := time.Now()
	for i := 0; i < max; i++ {
		log.Println("produce: Sending ", i)
		*p.active <- i
		time.Sleep(time.Second * 1)
	}
	TotalTime = time.Since(startTime)
	if sendMetrics{
		utils.SendTime(float64(TotalTime) / float64(time.Second),  utils.GetPath(monitoringURL, "updateTime"))
	}
	log.Println("produce: Done/ time taken: ", TotalTime)
	*p.done <- true // signal when done
	
}

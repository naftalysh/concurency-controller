package producer

import (
	"log"
	"time"
)

type Producer struct {
	Active *chan int
	Done *chan bool
}

var (
	TotalTime time.Duration
)

// NewProducer creates a Producer
func NewProducer(active *chan int, done *chan bool) *Producer {
	return &Producer{Active: active, Done: done}
}

func (p *Producer) ProduceInfinite(timeout time.Duration){
	log.Println("produce: Started")
	startTime := time.Now()
	var i int =0
	for  start := time.Now(); time.Since(start) < timeout; {
		log.Println("produce: Sending ", i)
		*p.Active <- i
		time.Sleep(time.Second * 1)
		i++
	}
	TotalTime = time.Since(startTime)
	log.Println("produce: Done/ time taken: ", TotalTime)
	*p.Done <- true // signal when done
}

func (p *Producer) Produce(max int) {
	log.Println("produce: Started")
	startTime := time.Now()
	for i := 0; i < max; i++ {
		log.Println("produce: Sending ", i)
		*p.Active <- i
		time.Sleep(time.Second * 1)
	}
	TotalTime = time.Since(startTime)
	log.Println("produce: Done/ time taken: ", TotalTime)
	*p.Done <- true // signal when done
	
}

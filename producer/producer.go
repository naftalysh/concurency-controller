package producer

import (
	"log"
	"sync"
	"time"
)

type Producer struct {
	WG *sync.WaitGroup
	Messages *chan int
	timeout time.Duration

}

type BatchProducer struct {
	WG *sync.WaitGroup
	Messages *chan int
	Batches int

}
var TotalTime time.Duration


func NewProducer(WG *sync.WaitGroup, Messages *chan int, timeout time.Duration) *Producer {
	return &Producer{
		WG: WG,
		Messages: Messages,
		timeout: timeout,
	}
}

func NewBatchProducer(WG *sync.WaitGroup, Messages *chan int, Batches int) *BatchProducer {
	return &BatchProducer{
		WG: WG,
		Messages: Messages,
		Batches: Batches,
	}
}

func (p *BatchProducer) Produce() {
	log.Println("Starting to produce messages")
	for i := 1; i <= p.Batches; i++ {
		*p.Messages <- i
		log.Println("Produced message:", i)
		time.Sleep(1 * time.Second)
	}
	close(*p.Messages)
	log.Println("Finished producing messages")
	p.WG.Done()
}

func (p *Producer) Produce() {
	log.Println("Starting to produce messages")
	count := 0
	for start := time.Now(); time.Since(start) < p.timeout; {
		*p.Messages <- count
		log.Println("Produced message:", count)
		time.Sleep(1 * time.Second)
		count++
	}
	close(*p.Messages)
	log.Println("Finished producing messages")
	p.WG.Done()
}





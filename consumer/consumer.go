package consumer

import (
	"log"
	"sync"
	"time"

	typesdef "github.com/redhat-appstudio-qe/concurency-controller/typesDef"
)


type Consumer struct {
	WG, WG2 *sync.WaitGroup
	Results *typesdef.Results
	Messages *chan int
	RPS int
}

var (
	AverageTimeForBatch time.Duration
	TotalRequestCopy int
)

func NewConsumer(WG *sync.WaitGroup, IWG *sync.WaitGroup, Messages *chan int, RPS int, results *typesdef.Results) *Consumer {
	return &Consumer{
		WG: WG,
		WG2: IWG,
		Messages: Messages,
		RPS: RPS,
		Results: results,
	}
}

func (c *Consumer) Consume(runner func()(error)) {
	log.Println("Starting to consume messages")
	failed := 0
	completed := 0
	total := 0
	for msg := range *c.Messages {
		log.Println("Received message:", msg)
		c.Results.RPS = c.RPS
		for i := 0; i < c.RPS; i++ {
			c.WG2.Add(1)
			total++
			c.Results.TotalRequests = total
			go c.Test(runner, &failed, &completed,i, msg)
			time.Sleep(30 * time.Microsecond)
		}
	}
	log.Println("Finished consuming messages")
	log.Println("Total: ", total)
	log.Println("Total failed: ", failed)
	c.Results.TotalRequests = total
	c.WG.Done()
}

func (c *Consumer) Test(runner func()(error), failed *int, completed *int, id int, batch int) (error) {
	log.Println("Starting test function: ", id, "for batch: ", batch)
	startTime := time.Now()
	if err := runner(); err != nil {
		*failed++
		c.Results.TotalErrors = *failed
		c.WG2.Done()
		return err
	}
	log.Println("Test function finished: ", id, "for batch: ", batch)
	log.Println("Total finished: ", *completed)
	Endtime := time.Since(startTime)
	AverageTimeForBatch += Endtime
	*completed++
	log.Println("Total finished: ", *completed)
	if AverageTimeForBatch != 0{
		c.Results.Latency = AverageTimeForBatch / time.Duration(*completed)
	}
	c.WG2.Done()
	return nil
}
package typesdef

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"time"
)

type Results struct {
	Latency time.Duration
	RPS int
	TotalRequests int
	TotalErrors int

}

type Consumer struct {
	Active *chan int
	Wg sync.WaitGroup
}


type RunnerFunction func()(error) 


func NewResults() *Results {
	return &Results{Latency: time.Duration(0), RPS: 0, TotalErrors: 0, TotalRequests: 0}
}

func PointerRunnerWrap(runner RunnerFunction ,i ...*int64)(error){
	var buffer bytes.Buffer
	buffer.WriteString("Making request: ")
	for _, val := range i {
		buffer.WriteString(fmt.Sprintf("%d ", *val))
	}
	log.Print(buffer.String())
	err := runner()
    *i[0]++
    if err != nil {
		return err
	}
    return nil
}

func RunnerWrap(runner RunnerFunction ,i ...int)(error){
	var buffer bytes.Buffer
	buffer.WriteString("Making request: ")
	for _, val := range i {
		buffer.WriteString(fmt.Sprintf("%d ", val))
	}
	log.Print(buffer.String())
	err := runner()
    if err != nil {
		return err
	}
    return nil
}
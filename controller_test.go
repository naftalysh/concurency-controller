package main

import (
	"errors"
	"log"
	"testing"
	"time"

	controller "github.com/redhat-appstudio-qe/concurency-controller/controller"
)

var Count int = 0

func testFunc() error {
    // Replace this with your actual function to test
	log.Println("Function Started!")
	time.Sleep(time.Second * 5)
	log.Println("Function Finished!")
	Count++
	if Count == 10{
		Count = 0
		return errors.New(`unexpected error`)
	}
    return nil 
}


func TestController(t *testing.T){
	// Execute in Batches 
	// This takes two parameters i.e max no of requests to make and number of batches it has to execute those requests 
	// RPS is automatically calculated based on the above params
	// example: assume MAX_REQ = 50 , BATCHES = 5 then RPS = 10 
	// if you want to capture/send metrics please  provide the third parameter i.e MonitoringURL
	// Monitoring URL should point to hosted/self hosted instance of https://github.com/redhat-appstudio-qe/perf-monitoring
	// if you dont want to push metrics then just pass an empty string
	/* MAX_REQ := 50
	BATCHES := 5
	result := controller.NewBatchController(MAX_REQ,BATCHES, "").ConcurentlyExecute(testFunc)
	log.Println("Result Array: ", result) */


	// Execute infinitely untill a timeout is met 
	// This takes two parameters i.e timeout duration and RPS (Requests Per Second)
	// if you want to capture/send metrics please  provide the third parameter i.e MonitoringURL
	// Monitoring URL should point to hosted/self hosted instance of https://github.com/redhat-appstudio-qe/perf-monitoring
	// if you dont want to push metrics then just pass an empty string
	// TIMEOUT := 10 * time.Second
	// RPS := 3
	// result := controller.NewInfiniteController(RPS, TIMEOUT, "").ConcurentlyExecuteInfinite(testFunc)
	// log.Println("Result Array: ", result)


	// Executes infinitely untill a timeout is met 
	// This takes three parameters i.e timeout duration and MAX_RPS (Requests Per Second) and error threshold 
	// Runs in a way like locust it keeps on callinf the runner function, if the execution is error free it doulbles the RPS
	// if the execution has errors then it decrements the RPS gives more info on latency
	// if you want to capture/send metrics please  provide the third parameter i.e MonitoringURL
	// Monitoring URL should point to hosted/self hosted instance of https://github.com/redhat-appstudio-qe/perf-monitoring
	// if you dont want to push metrics then just pass an empty string
	TIMEOUT := 20 * time.Second
	maxRPS := 50
	errorThresholdRate := 0.5
	result := controller.NewSpikeController(maxRPS, TIMEOUT, errorThresholdRate, "").CuncurentlyExecuteSpike(testFunc)
	log.Println("Result Array: ", result)
}




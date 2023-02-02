package main

import (
	"log"
	"testing"
	"time"

	controller "github.com/redhat-appstudio-qe/concurency-controller/controller"
)


func testFunction() error {
    // Replace this with your actual function to test
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
	MAX_REQ := 50
	BATCHES := 5
	result := controller.NewLoadController(MAX_REQ,BATCHES, "").ConcurentlyExecute(testFunction)
	log.Println("Result Array: ", result)


	// Execute infinitely untill a timeout is met 
	// This takes two parameters i.e timeout duration and RPS (Requests Per Second)
	// if you want to capture/send metrics please  provide the third parameter i.e MonitoringURL
	// Monitoring URL should point to hosted/self hosted instance of https://github.com/redhat-appstudio-qe/perf-monitoring
	// if you dont want to push metrics then just pass an empty string
	TIMEOUT := 60 * time.Second
	RPS := 3
	result = controller.NewInfiniteLoadController(TIMEOUT, RPS, "").ConcurentlyExecuteInfinite(testFunction)
	log.Println("Result Array: ", result)


	// Executes infinitely untill a timeout is met 
	// This takes three parameters i.e timeout duration and MAX_RPS (Requests Per Second) and error threshold 
	// Runs in a way like locust it keeps on callinf the runner function, if the execution is error free it doulbles the RPS
	// if the execution has errors then it decrements the RPS gives more info on latency
	// if you want to capture/send metrics please  provide the third parameter i.e MonitoringURL
	// Monitoring URL should point to hosted/self hosted instance of https://github.com/redhat-appstudio-qe/perf-monitoring
	// if you dont want to push metrics then just pass an empty string
	//TIMEOUT := 60 * time.Second
	maxRPS := 50
	errorThresholdRate := 0.5
	result = controller.NewSpikeLoadController(TIMEOUT, maxRPS, errorThresholdRate, "").CuncurentlyExecuteSpike(testFunction)
	log.Println("Result Array: ", result)
}




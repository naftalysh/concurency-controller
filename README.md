
# Concurency Controller

[![Go](https://github.com/redhat-appstudio-qe/concurency-controller/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/redhat-appstudio-qe/concurency-controller/actions/workflows/go.yml)

**The "concurency-controller" is a Go based concurrent requests controller, consisting of three controllers: Load Controller, Infinite Load Controller, and Spike Load Controller.**

- The **Load Controller** takes two arguments: `MAX_REQ` (the number of requests to be made) and `BATCHES` (the number of threads to execute). **The `RPS` is calculated by dividing `MAX_REQ` by `BATCHES` and the controller makes a number of concurrent requests each second equal to the RPS, until `MAX_REQ` has been made.**

- The **Infinite Load Controller** takes two arguments: `timeout` and `RPS`. **It will continuously make concurrent requests based on the set RPS until the timeout is reached.**

- The **Spike Load Controller** takes three arguments: `timeout`, `maxRPS`, and `errorThresholdrate`. **It starts making requests based on a minimum `RPS` of 1, and if the previous 10 requests are error-free, it will increase the `RPS by a factor of 2`. If there are errors, the error rate will be checked, and if it crosses the threshold, the RPS will be decremented to the previously worked RPS.**

- Each of the three controllers in the "concurency-controller" repo also takes an optional argument, which is the URL for a running instance of [perf-monitoring](https://github.com/redhat-appstudio-qe/perf-monitoring), a part of the monitoring stack that ingests the metrics captured by these three controllers. If you do not want to use the monitoring, simply leave the argument as an empty string.


**Each of the controllers has its own functions for execution, i.e  `ConcurrentlyExecute` for LoadController, `ConcurrentlyExecuteInfinite` for InfiniteLoadController, and `ConcurrentlyExecuteSpike` for SpikeLoadController.** These functions return a result array of type `Results`, which is a struct that contains the following fields:

- **Latency**: the time duration of the request
- **RPS**: the number of requests made per second
- **TotalRequests**: the total number of requests made
- **TotalErrors**: the total number of errors encountered during the execution of the requests.

## Installation

Use this package in your go code 

```bash
  go get github.com/redhat-appstudio-qe/concurency-controller
```
    

### Commons 

#### The inititlize and finish functions 

* **Initilize : The function is initializing a LoadController struct. The LoadController struct is a custom data structure used to control the load on some system. The function performs the following steps:**

    * It sets the value of the "Results" field of the LoadController struct to a new instance of the "typesdef.NewResults()" type. The "Results" field is a field of the LoadController struct that stores the results of load testing.

    * It creates a new channel called "StopTicker" of type "chan struct{}". The "StopTicker" channel is used to stop some ticker that's running in the background.

    * The "sendMetrics" field is set to true by default. If the "monitoringURL" field is an empty string, then "sendMetrics" is set to false. The "sendMetrics" field is used to determine if metrics should be sent to some monitoring service.

    * The "cpuprofile" and "memprofile" fields are either assigned the value of a flag if it exists, or they are created as new flags. The "cpuprofile" and "memprofile" fields are used to profile the CPU and memory usage of the program, respectively.

    * If "sendMetrics" is true, a new instance of the "utils.NewUtils()" type is created and assigned to the "util" field. The "SaveMetrics()" function is then called with "Results" as an argument. The "util" field is used to send metrics to some monitoring service.

    * The "RPS" (Requests Per Second) field is calculated as the result of dividing the "MaxReq" field by the "Batches" field. The "RPS" field is used to control the rate of requests per second to some system. If "RPS" is less than 1, a panic is raised with the message "Zero Requests Per Second Detected!". This is to ensure that the rate of requests is not set to 0.

    * The number of cores available is determined and "runtime.GOMAXPROCS()" is called to set the maximum number of cores to use. This is to optimize the program for performance by utilizing multiple cores if available.

    * Finally, if the "cpuprofile" flag is not an empty string, a CPU profile is started, and it will stop when the function ends. The CPU profile is used to analyze the program's CPU usage and identify performance bottlenecks.

In summary, this function initializes the LoadController struct by setting its fields, calculating the "RPS" field, optimizing the program for performance, and starting a CPU profile if the "cpuprofile" flag is not an empty string.

* **Finish : The purpose of this function is to clean up and prepare the system for shutting down. Here is a more detailed explanation of each step in the function:**

    * close(StopTicker): This line closes the StopTicker channel, which is used for gathering metrics during the load test. By closing this channel, the process of gathering metrics will stop.

    * if l.sendMetrics: If the sendMetrics flag is set to true, the function logs all collected metrics stored in the ResultsArray variable. These metrics can include information such as the number of requests per second, average latency, and other performance metrics, the function resets the RPS and Latency fields of the Results struct to zero. This is done to prepare for the next load test. Finally, the function calls l.SaveMetrics to save the metrics.

    * if *l.memprofile != "": If the memprofile variable is not an empty string, the function creates a memory profile file. Memory profiling is a technique used to analyze the memory usage of a program. This step creates a file to store the profile. The name of the file is specified by the value of l.memprofile.

    * runtime.GC(): This line triggers a garbage collection operation in the Go runtime. Garbage collection is a mechanism in Go that automatically frees memory that is no longer in use by the program.

    * pprof.WriteHeapProfile(f): This line writes the heap profile to the file created in step 3. A heap profile shows the memory usage of a Go program at a particular moment in time, including information about the size of allocated objects, the number of objects, and the types of objects.

### LoadController & InfiniteLoadController

#### Approach Explanation

- ***producer-consumer*** pattern. The `Producer` generates a specified number of `active threads` and sends them over a channel. The `Consumer` receives these `active thread`s, runs a specified number of `requests per second` for each active thread, and calculates the **total number of requests, successful requests, failed requests, and average latency.** The results are logged and, if specified, sent to a monitoring URL. The Producer can produce an infinite number of batches until a timeout is reached or a specified number of batches are generated.

* **In Depth Explanation** 
    * Consumer : A `Consumer` struct is defined, which has an `Active` channel and a `Results` struct. The Consume method is used to read messages from the `Active` channel and to perform the actual concurrency testing. The function calls the `CommonConsume` method, which receives the `RPS (requests per second)`, `the runner function`, `the number of batches`, the `monitoring URL`, and a flag indicating whether metrics should be sent.
        * The Consume method starts a loop that waits for a message from the `Active` channel. Upon receiving a message, the loop starts a number of `goroutines` equal to the `RPS` value passed in. Each goroutine calls the `runner` function and runs it with the `id` of the loop and the `active` thread value. If the runner function returns an error, it increments the `TotalFailedReq` counter. After starting all the goroutines, it waits for them to complete using the `WaitGroup` and calculates the **AverageTimeForBatch, TotalRequests and TotalErrors** Finally, it prints the metrics using the `metricsPrinter` method.
    * Producer: A Producer struct has an `Active` channel and a `Done` channel. The `Produce` method is used to send messages to the `Active channel`, with the number of messages being either ***infinite (using ProduceInfinite)*** or a ***specified maximum (using Produce).*** The `Done` channel is used to signal when all messages have been sent.
        * The `Produce` method starts a loop that sends messages to the `Active `channel and sleeps for a second before sending the next message. If a `maximum number of iterations` is provided, the loop will stop after reaching that number. If a `timeout`is provided, the loop will continue for the duration of the timeout. After the loop is done, the TotalTime is calculated and a message is sent to the `Done` channel to signal that it is complete.
    * In short, the code creates a concurrency control mechanism using channels and goroutines. The Producer sends messages to the Consumer who processes them and prints metrics. The code demonstrates the use of WaitGroup and channels to control concurrency and communication between Goroutines.

    #### **Execute a LoadTest with LoadController**

    ```golang
    package main

    import (
        "log"
        "time"
        controller "github.com/redhat-appstudio-qe/concurency-controller/controller"
    )

    func testFunction() error {
        // Replace this with your actual function to test
        return nil
    }

    // Execute in Batches 
    // This takes two parameters i.e max no of requests to make and number of batches it has to execute those requests 
    // RPS is automatically calculated based on the above params
    // example: assume MAX_REQ = 50 , BATCHES = 5 then RPS = MAX_REQ/BATCHES = 10 
    // if you want to capture/send metrics please  provide the third parameter i.e MonitoringURL
    // Monitoring URL should point to hosted/self hosted instance of https://github.com/redhat-appstudio-qe/perf-monitoring
    // if you dont want to push metrics then just pass an empty string
    func main(){
        MAX_REQ := 50
        BATCHES := 5
        r := controller.NewLoadController(MAX_REQ,BATCHES, "").ConcurentlyExecute(testFunction)
        log.Println(r)
    }

    ```

    #### **Execute a LoadTest with InfiniteLoadController**

    Make the following changes to the above golang code 

    ```golang
    // Execute infinitely untill a timeout is met based on RPS specified
    // This takes two parameters i.e timeout duration and RPS (Requests Per Second)
    func main(){
        TIMEOUT := 60 * time.Second
        RPS := 3
        r := controller.NewInfiniteLoadController(TIMEOUT, RPS, "").ConcurentlyExecuteInfinite(testFunction)
        log.Println(r)
    }

    ```
### SpikeLoadController

#### Approach Explanation

*  **This Controller implements a load test using a spike strategy.** tests the performance of a system under different loads. The load controller creates a number of concurrent processes (Goroutines) that run a test function at the same time, and it adjusts the number of requests per second (RPS) over time. The goal is to find the maximum RPS that can be handled by the system without exceeding an error threshold, which is specified as a rate (error count divided by total requests).

* The function sets a timer to run for a specified timeout duration and stops the test when the timer triggers. It calculates various performance metrics during the test, including RPS, error count, average latency (the total latency divided by the number of requests), max latency, and min latency. It adjusts the RPS based on the error rate, reducing the RPS if the error rate exceeds the error threshold and increasing it otherwise. The results of the test are logged and stored in the Results field of the LoadController struct.

* The load controller uses the Go sync package to manage the concurrency, and it uses the time package to measure elapsed time and set the timer. It also uses the log package to log information about the test results.

* **Here is how it works**:

    * ___Variables declaration___:
        * `wg`: a WaitGroup used to wait for all Goroutines to complete.
        * `start`: a time.Time variable that stores the start time of the test.
        * `rps`: an integer variable that stores the current RPS (Requests per second) rate. The initial value is 1.
        * `maxRPS`: an integer variable that stores the maximum RPS rate allowed.
        * `errorThreshold`: a float64 variable that stores the error threshold, beyond which the RPS rate will be reduced.
        * `errorCount`: an integer variable that stores the number of errors that occurred during the test.
        * `totalRequests`: an integer variable that stores the total number of requests made during the test.
        * `totalLatency`: a time.Duration variable that stores the total latency of all requests.
        * `maxLatency`: a time.Duration variable that stores the maximum latency of all requests.
        * `minLatency`: a time.Duration variable that stores the minimum latency of all requests.
        * `timeout`: a time.Duration variable that stores the timeout duration for the test.
        * `successCount`: an integer variable that stores the number of consecutive successful requests.
    * `Timer and Select statement`: The code sets up a timer to check if the test has timed out after the specified duration in timeout. If the test has timed out, the code logs the results, including the total number of requests made, error count, average latency, maximum latency, and minimum latency, and returns. If the test is not timed out, the code proceeds to the next step.

    * `Executing the test function`: The code adds a new Goroutine to the WaitGroup, which will execute the test function. The test function is executed by calling the typesdef.PointerRunnerWrap function and passing the test function and a counter variable as arguments. The start time of the request is recorded before executing the test function and the latency of the request is calculated by subtracting the start time from the current time after the execution. The latency is then accumulated in the totalLatency variable and compared with the current maxLatency and minLatency to update them if necessary. If an error occurs during the test function execution, the error count is incremented.

    * `Measuring RPS and updating results`: After each batch of requests, the code calculates the RPS by dividing the total number of requests by the duration of the test since the start time. The RPS is then logged. The code updates the results structure by updating the RPS, total number of requests, total number of errors, and average latency.

    * `Adjusting RPS rate`: The code calculates the error rate by dividing the error count by the total number of requests. If the error rate is greater than the specified error threshold, the RPS is reduced to half the current value. If the error rate is less than the threshold, the code increments the success count. If the success count reaches 10, the RPS is increased to double the current value but limited by the maxRPS value. The code waits for all Goroutines

    #### **Execute a LoadTest with SpikeLoadController**

    Make the following changes to the above golang code 

    ```golang
    // Executes infinitely untill a timeout is met 
	// This takes three parameters i.e timeout duration and MAX_RPS (Requests Per Second) and error threshold 
	// Runs in a way like locust it keeps on callinf the runner function, if the execution is error free it doulbles the RPS
	// if the execution has errors then it decrements the RPS gives more info on latency
    func main(){
        TIMEOUT := 60 * time.Second
        maxRPS := 50
        errorThresholdRate := 0.5
        r := controller.NewSpikeLoadController(TIMEOUT, maxRPS, errorThresholdRate, "").CuncurentlyExecuteSpike(testFunction)
        log.Println(r)
    }

    ```








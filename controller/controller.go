package controller

import (
	"flag"
	//"sync"
	//"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	consumer "github.com/redhat-appstudio-qe/concurency-controller/consumer"
	producer "github.com/redhat-appstudio-qe/concurency-controller/producer"
	typesdef "github.com/redhat-appstudio-qe/concurency-controller/typesDef"
	"github.com/redhat-appstudio-qe/concurency-controller/utils"
)

var (
	ResultsArray []typesdef.Results
	StopTicker chan struct{}
	
)

type LoadController struct{
	MaxReq, Batches, RPS, maxRPS int
	errorThreshold float64
	cpuprofile, memprofile  *string
	timeout time.Duration
	sendMetrics bool
	monitoringURL string
	Results *typesdef.Results
	util *utils.UtilsType
}
type Runner func()(error)

// Divide takes two integers, M and B, divides them and returns an integer.
// Panics on Error
func divide(M int, B int) int {
	if M < 1 || B < 1 {
		panic("invalid args!")
	}
	return M / B

}

// `NewLoadController` creates a new LoadController struct with the given parameters
func NewLoadController(MaxReq int, Batches int, MonitoringURL string) *LoadController {
	return &LoadController{MaxReq: MaxReq, Batches: Batches, monitoringURL: MonitoringURL}
}

// `NewInfiniteLoadController` creates a new LoadController object with the given timeout, RPS, and MonitoringURL
func NewInfiniteLoadController(timeout time.Duration, RPS int, MonitoringURL string) *LoadController {
	return &LoadController{timeout: timeout, RPS: RPS, monitoringURL: MonitoringURL}
}

// `NewSpikeLoadController` creates a new `LoadController` with the given `timeout`, `maxRPS`,
// `errorThreshold` and `MonitoringURL`
func NewSpikeLoadController(timeout time.Duration, maxRPS int, errorThreshold float64, MonitoringURL string) *LoadController {
	return &LoadController{timeout: timeout, monitoringURL: MonitoringURL, maxRPS: maxRPS, errorThreshold: errorThreshold}
}

// Initializing the LoadController struct.
func (l *LoadController) initialize(infinite bool){
	l.Results = typesdef.NewResults();
	StopTicker = make(chan struct{})
	
	l.sendMetrics = true;
	if l.monitoringURL == "" {
		l.sendMetrics = false
	}
	log.Println("Sending metrics set to:", l.sendMetrics, "|", l.monitoringURL)

	if flag.Lookup("cpuprofile") == nil && flag.Lookup("memprofile") == nil {
		l.cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
		l.memprofile = flag.String("memprofile", "", "write memory profile to `file`")
		flag.Parse()
	} else {
		l.cpuprofile = &flag.Lookup("cpuprofile").DefValue
		l.memprofile = &flag.Lookup("memprofile").DefValue
	}

	if l.sendMetrics{
		l.util = utils.NewUtils(l.monitoringURL)
		l.SaveMetrics(l.Results)
		log.Println("Reset Metrics in Push gateway to zero and Sleeping 5 seconds")

		time.Sleep(5 * time.Second)
	}

	
	if !infinite{
		l.RPS = divide(l.MaxReq, l.Batches)
		if l.RPS < 1 {
			panic("Zero Requests Per Second Detected!")
		}
	}
    
	// utilize the max num of cores available
	runtime.GOMAXPROCS(runtime.NumCPU())

	// CPU Profile
	if *l.cpuprofile != "" {
		f, err := os.Create(*l.cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}
	

}

// A function that is called at the end of the program. It stops the ticker, and it writes the memory
// profile to a file.
func (l *LoadController) finish(){

	//stop gathering metrics 
	close(StopTicker)

	//send Metrics 
	if l.sendMetrics{
		log.Println("All Metrics Collected: ", ResultsArray)
	}

	// Memory Profile
	if *l.memprofile != "" {
		f, err := os.Create(*l.memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}

	//reset necessary metrics 
	if l.sendMetrics{
		l.Results.RPS = 0 
		l.Results.Latency = 0
		l.SaveMetrics(l.Results)
	}
}

// Defining a function that takes a runner function as an argument and returns a slice of Results.
func (l *LoadController) ConcurentlyExecuteInfinite(runner typesdef.RunnerFunction) []typesdef.Results {
	l.initialize(true)
	// start metrics
	var active = make(chan int)  // channel to send messages
	var done = make(chan bool) // channel to control when production is done

	// Start a goroutine for Produce.produce
	// Start a goroutine for Consumer.consume
	l.SaveMetricsOnTick(l.Results)
	go consumer.NewConsumer(&active, l.Results).Consume(l.RPS, runner, l.Batches, l.monitoringURL, l.sendMetrics)
	go producer.NewProducer(&active, &done).ProduceInfinite(l.timeout)

	// Finish the program when the production is done
	<-done

	l.finish()
	return ResultsArray
}


// A function that takes a runner function as an argument and returns a slice of Results.
func  (l *LoadController) ConcurentlyExecute(runner typesdef.RunnerFunction) []typesdef.Results {
	
	l.initialize(false)

	// start metrics
	var active = make(chan int)  // channel to send messages
	var done = make(chan bool) // channel to control when production is done

	// Start a goroutine for Produce.produce
	// Start a goroutine for Consumer.consume
	l.SaveMetricsOnTick(l.Results)
	go consumer.NewConsumer(&active, l.Results).Consume(l.RPS, runner, l.Batches, l.monitoringURL, l.sendMetrics)
	go producer.NewProducer(&active, &done).Produce(l.Batches)

	// Finish the program when the production is done
	<-done

	l.finish()
	return ResultsArray

}

// A function that takes a runner function as an argument and returns a slice of Results.
func  (l *LoadController) CuncurentlyExecuteSpike(runner typesdef.RunnerFunction) []typesdef.Results {
	
	l.initialize(true)
	log.Println(l.timeout,l.maxRPS)
	l.SaveMetricsOnTick(l.Results)
	l.ExecuteSpike(runner)
	l.SaveMetrics(l.Results)
	l.finish()
	return ResultsArray
	
}

// It creates a new ticker that ticks every second, and then it starts a goroutine that listens for the
// ticker to tick, and when it does, it calls the SaveMetrics function, and when it receives a message
// on the StopTicker channel, it stops the ticker and returns
func (l *LoadController) SaveMetricsOnTick(R *typesdef.Results){
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
		   select {
			case <- ticker.C:
				l.SaveMetrics(R)
			case <- StopTicker:
				ticker.Stop()
				return
			}
		}
	 }()

}

// The function takes a pointer to a struct of type Results, and appends it to a slice of type Results
func (l *LoadController) SaveMetrics(R *typesdef.Results) ([]typesdef.Results){
	ResultsArray = append(ResultsArray, *R)
	if l.sendMetrics{
		l.util.SendMetrics(R.TotalRequests, R.TotalErrors, R.Latency, R.RPS)
	}
	return ResultsArray
}

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
	"github.com/redhat-appstudio-qe/concurency-controller/utils"
)

var (
	ResultsArray []Results
	StopTicker chan struct{}
	
)


type Results struct {
	Latency time.Duration
	RPS int
	TotalRequests int
	TotalErrors int

}

type LoadController struct{
	MaxReq, Batches, RPS, maxRPS int
	errorThreshold float64
	cpuprofile, memprofile  *string
	timeout time.Duration
	sendMetrics bool
	monitoringURL string
	Results *Results
}
type Runner func()(error)

func divide(M int, B int) int {
	if M < 1 || B < 1 {
		panic("invalid args!")
	}
	return M / B

}

func NewResults() *Results {
	return &Results{Latency: time.Duration(0), RPS: 0, TotalErrors: 0, TotalRequests: 0}
}

// NewConsumer creates a Consumer
func NewLoadController(MaxReq int, Batches int, MonitoringURL string) *LoadController {
	return &LoadController{MaxReq: MaxReq, Batches: Batches, monitoringURL: MonitoringURL}
}

func NewInfiniteLoadController(timeout time.Duration, RPS int, MonitoringURL string) *LoadController {
	return &LoadController{timeout: timeout, RPS: RPS, monitoringURL: MonitoringURL}
}

func NewSpikeLoadController(timeout time.Duration, maxRPS int, errorThreshold float64,MonitoringURL string) *LoadController {
	return &LoadController{timeout: timeout, monitoringURL: MonitoringURL, maxRPS: maxRPS, errorThreshold: errorThreshold}
}

func (l *LoadController) initialize(infinite bool){
	l.Results = NewResults();
	StopTicker = make(chan struct{})
	l.sendMetrics = false;

	if flag.Lookup("cpuprofile") == nil && flag.Lookup("memprofile") == nil && flag.Lookup("monitoringURL") == nil{
		l.cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
		l.memprofile = flag.String("memprofile", "", "write memory profile to `file`")
		flag.Parse()
	} else {
		l.cpuprofile = &flag.Lookup("cpuprofile").DefValue
		l.memprofile = &flag.Lookup("memprofile").DefValue
	}

	if l.monitoringURL != ""{
		l.sendMetrics = true
	}

	log.Println("Send metrics:", l.sendMetrics, "|", l.monitoringURL)
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

func (l *LoadController) finish(){

	//stop gathering metrics 
	close(StopTicker)

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


}

func (l *LoadController) ExecuteInfinite(runner utils.RunnerFunction){
	l.initialize(true)
	// start metrics
	var active = make(chan int)  // channel to send messages
	var done = make(chan bool) // channel to control when production is done

	// Start a goroutine for Produce.produce
	// Start a goroutine for Consumer.consume
	go consumer.NewConsumer(&active).Consume(l.RPS, runner, l.Batches, l.monitoringURL, l.sendMetrics)
	go producer.NewProducer(&active, &done).ProduceInfinite(l.timeout, l.monitoringURL, l.sendMetrics)

	// Finish the program when the production is done
	<-done

	l.finish()
}


func  (l *LoadController) ConcurentlyExecute(runner utils.RunnerFunction){
	
	l.initialize(false)

	// start metrics
	var active = make(chan int)  // channel to send messages
	var done = make(chan bool) // channel to control when production is done

	// Start a goroutine for Produce.produce
	// Start a goroutine for Consumer.consume
	go consumer.NewConsumer(&active).Consume(l.RPS, runner, l.Batches, l.monitoringURL, l.sendMetrics)
	go producer.NewProducer(&active, &done).Produce(l.Batches, l.monitoringURL, l.sendMetrics)

	// Finish the program when the production is done
	<-done

	l.finish()

}

func  (l *LoadController) CuncurentSpikeExecutor(runner utils.RunnerFunction) []Results {
	
	l.initialize(true)
	SaveMetricsOnTick(l.Results)
	l.ExecuteSpike(runner)
	SaveMetrics(l.Results)
	l.finish()
	return ResultsArray
	
}

func SaveMetricsOnTick(R *Results){
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
		   select {
			case <- ticker.C:
				SaveMetrics(R)
			case <- StopTicker:
				ticker.Stop()
				return
			}
		}
	 }()

}

func SaveMetrics(R *Results) ([]Results){
	ResultsArray = append(ResultsArray, *R)
	return ResultsArray
}

package controller

import (
	"flag"
	"sync"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/redhat-appstudio-qe/concurency-controller/consumer"
	"github.com/redhat-appstudio-qe/concurency-controller/producer"
	"github.com/redhat-appstudio-qe/concurency-controller/spiker"
	typesdef "github.com/redhat-appstudio-qe/concurency-controller/typesDef"
	"github.com/redhat-appstudio-qe/concurency-controller/utils"
)

var (
	ResultsArray []typesdef.Results
	StopTicker chan struct{}
	
)

type LoadType struct{
	cpuprofile, memprofile  *string
	sendMetrics bool
	RPS int
	monitoringURL string
	Results *typesdef.Results
	util *utils.UtilsType
}

type BatchLoadType struct {
	LoadType
	MaxReq, Batches int
}

type InfiniteLoadType struct {
	LoadType
	timeout time.Duration
	RPS int
}

type SpikeLoadType struct {
	LoadType
	timeout time.Duration
	errorThreshold float64
	maxRPS int
}

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

func NewBatchController(MaxReq int, Batches int, MonitoringURL string) *BatchLoadType{
	LoadTypeVar := LoadType{monitoringURL: MonitoringURL}
	return &BatchLoadType{LoadType : LoadTypeVar, MaxReq: MaxReq, Batches: Batches}
}

func NewInfiniteController(RPS int, timeout time.Duration, MonitoringURL string) *InfiniteLoadType{
	LoadTypeVar := LoadType{monitoringURL: MonitoringURL}
	return &InfiniteLoadType{LoadType : LoadTypeVar, RPS: RPS, timeout: timeout}
}

func NewSpikeController(maxRPS int, timeout time.Duration, errorThreshold float64, MonitoringURL string) *SpikeLoadType{
	LoadTypeVar := LoadType{monitoringURL: MonitoringURL}
	return &SpikeLoadType{LoadType : LoadTypeVar, maxRPS: maxRPS, timeout: timeout, errorThreshold:  errorThreshold}
}


func(l *LoadType) commonInit() {
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

func (l *LoadType) commonfinish(){

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


func (l *BatchLoadType) init(){
	l.RPS = divide(l.MaxReq, l.Batches)
	if l.RPS < 1 {
		panic("Zero Requests Per Second Detected!")
	}
}



func (l *InfiniteLoadType) ConcurentlyExecuteInfinite(runner typesdef.RunnerFunction)[]typesdef.Results {

	log.Println("Main function started")
	l.LoadType.commonInit()

	var wg, wg2 sync.WaitGroup
	wg.Add(2)
	messages := make(chan int)
	l.SaveMetricsOnTick(l.Results)
	go producer.NewProducer(&wg, &messages, l.timeout).Produce()
	go consumer.NewConsumer(&wg, &wg2, &messages, l.RPS, l.Results).Consume(runner)

	log.Println("Main function continuing")
	wg.Wait()
	wg2.Wait()
	log.Println("STATS: ", l.Results)
	log.Println("Main function finished")


	l.LoadType.commonfinish()

	return ResultsArray

}

// A function that takes a runner function as an argument and returns a slice of Results.
func  (l *BatchLoadType) ConcurentlyExecute(runner typesdef.RunnerFunction) []typesdef.Results {
	
	log.Println("Main function started")
	l.LoadType.commonInit()
	l.init()

	var wg, wg2 sync.WaitGroup
	wg.Add(2)
	messages := make(chan int)
	l.SaveMetricsOnTick(l.Results)
	go producer.NewBatchProducer(&wg, &messages, l.Batches).Produce()
	go consumer.NewConsumer(&wg, &wg2, &messages, l.RPS, l.Results).Consume(runner)

	log.Println("Main function continuing")
	wg.Wait()
	wg2.Wait()
	log.Println("STATS: ", l.Results)
	log.Println("Main function finished")
	l.LoadType.commonfinish()

	return ResultsArray

}

// A function that takes a runner function as an argument and returns a slice of Results.
func  (l *SpikeLoadType) CuncurentlyExecuteSpike(runner typesdef.RunnerFunction) []typesdef.Results {
	
	log.Println("Main function started")
	l.LoadType.commonInit()
	l.SaveMetricsOnTick(l.Results)
	spiker.NewSpiker(l.maxRPS, l.timeout, l.errorThreshold, l.Results).GenerateSpike(runner)
	log.Println("Main function continuing")
	l.LoadType.commonfinish()
	log.Println("Main function finished")

	return ResultsArray
	
}

// It creates a new ticker that ticks every second, and then it starts a goroutine that listens for the
// ticker to tick, and when it does, it calls the SaveMetrics function, and when it receives a message
// on the StopTicker channel, it stops the ticker and returns
func (l *LoadType) SaveMetricsOnTick(R *typesdef.Results){
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
func (l *LoadType) SaveMetrics(R *typesdef.Results) ([]typesdef.Results){
	ResultsArray = append(ResultsArray, *R)
	if l.sendMetrics{
		l.util.SendMetrics(R.TotalRequests, R.TotalErrors, R.Latency, R.RPS)
	}
	return ResultsArray
}

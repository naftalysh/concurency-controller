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
)

type LoadController struct{
	MaxReq, Batches, RPS, maxRPS int
	errorThreshold float64
	cpuprofile, memprofile  *string
	timeout time.Duration
	sendMetrics bool
	monitoringURL string
}
type Runner func(...int)(error)

func divide(M int, B int) int {
	if M < 1 || B < 1 {
		panic("invalid args!")
	}
	return M / B

}

// NewConsumer creates a Consumer
func NewLoadController(MaxReq int, Batches int, MonitoringURL string) *LoadController {
	return &LoadController{MaxReq: MaxReq, Batches: Batches, monitoringURL: MonitoringURL}
}

func NewInfiniteLoadController(timeout time.Duration, RPS int, MonitoringURL string) *LoadController {
	return &LoadController{timeout: timeout, RPS: RPS, monitoringURL: MonitoringURL}
}

func NewSpikeLoadController(timeout time.Duration, maxRPS int, errorThreshold float64, MonitoringURL string) *LoadController {
	return &LoadController{timeout: timeout, monitoringURL: MonitoringURL, maxRPS: maxRPS, errorThreshold: errorThreshold}
}

func (l *LoadController) initialize(infinite bool){
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

func (l *LoadController) ExecuteInfinite(runner consumer.ConsumerFunction){
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


func  (l *LoadController) ConcurentlyExecute(runner consumer.ConsumerFunction){
	
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

func  (l *LoadController) CuncurentSpikeExecutor(runner RunnerFunction){
	l.initialize(true)
	l.ExecuteSpike(runner)
	l.finish()
	
}

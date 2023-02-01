package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

)

var (
	PushMetricsPath string = "/pushMetrics"
	HealthCheckPath string = "/"
)

type UtilsType struct{
	ExporterURL string
}

func NewUtils(ExporterURL string) *UtilsType {
	return &UtilsType{ExporterURL: ExporterURL}
}

func GetPath(URL, sub string) string{
	return URL + sub
}

func (u UtilsType) CheckExporterConnection(){
	resp, err := http.Get(GetPath(u.ExporterURL, HealthCheckPath))
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	log.Println("Connect to Metrics Exporter status: ", resp.StatusCode)
	if resp.StatusCode != 200 {
		log.Fatalf("An unknown error occured")
	}
}

func (u UtilsType) SendMetrics(Total int, Failed int, Latency time.Duration, RPS int){
	
	postBody, _ := json.Marshal(map[string]float64{
		"total":  float64(Total),
		"failed": float64(Failed),
		"latency": float64(Latency/time.Microsecond),
		"RPS": float64(RPS),
	 })
	 responseBody := bytes.NewBuffer(postBody)
  	//Leverage Go's HTTP Post function to make request
	 resp, err := http.Post(GetPath(u.ExporterURL, PushMetricsPath), "application/json", responseBody)
  	//Handle Error
	 if err != nil {
		log.Fatalf("An Error Occured %v", err)
	 }
	 defer resp.Body.Close()
  	//Read the response body
	 body, err := io.ReadAll(resp.Body)
	 if err != nil {
		log.Fatalln(err)
	 }
	 sb := string(body)
	 log.Println(sb)
}

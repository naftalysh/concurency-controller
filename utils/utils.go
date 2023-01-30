package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func GetPath(URL, sub string) string{
	return URL + sub
}

func SendMetrics(T float64, S float64, F float64, path string){
	
	postBody, _ := json.Marshal(map[string]float64{
		"total":  T,
		"failed": F,
		"success": S,
	 })
	 responseBody := bytes.NewBuffer(postBody)
  	//Leverage Go's HTTP Post function to make request
	 resp, err := http.Post(path, "application/json", responseBody)
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

func SendTotal(T float64, path string){
	
	postBody, _ := json.Marshal(map[string]float64{
		"totalReq":  T,
	 })
	 responseBody := bytes.NewBuffer(postBody)
  	//Leverage Go's HTTP Post function to make request
	 resp, err := http.Post(path, "application/json", responseBody)
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

func SendTime(T float64, path string){
	postBody, _ := json.Marshal(map[string]float64{
		"time":  T,
	 })
	 responseBody := bytes.NewBuffer(postBody)
  	//Leverage Go's HTTP Post function to make request
	 resp, err := http.Post(path, "application/json", responseBody)
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
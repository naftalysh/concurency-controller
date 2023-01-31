package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)
type RunnerFunction func()(error) 

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
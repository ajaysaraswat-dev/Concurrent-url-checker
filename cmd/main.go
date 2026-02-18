package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ajaysaraswat/concurrency/internal/config"
)


type Result struct {
	Url string
	Status int
	Error string

}




func main(){
	cfg,err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal("error to load config file",err)
	}
	

	workerCount := cfg.Server.Workers
	queueSize := cfg.Server.QueueSize
	Timeout := cfg.Server.Timeout


	inputfile, err := os.Open(cfg.InputFile)
    if err != nil {
	log.Fatalf("failed to open input file: %v", err)
    }   

	//create the output file 
	outputfile,err := os.Create(cfg.OutputFile)
	if err != nil {
		panic(err)
	}
	defer outputfile.Close()
	//created the queues
	jobs := make(chan string,queueSize)
	results := make(chan Result,queueSize)

	client := &http.Client{
		Timeout: time.Duration(Timeout)*time.Second,
	}
	var wg sync.WaitGroup

	//start workers 
	for i:=0;i<workerCount;i++ {
		wg.Add(1)
		go processUrl(i,jobs,results,client,&wg)

	}
	go Writer(results,outputfile)

	scanner := bufio.NewScanner(inputfile)
	for scanner.Scan() {
		jobs <- scanner.Text()
	}

	close(jobs)

	wg.Wait()

	close(results)

	fmt.Println("all urls has processed succesfully")
	
}

func processUrl(i int,jobs <-chan string,results chan <-Result,client *http.Client,wg *sync.WaitGroup){
	defer wg.Done()
	for url := range jobs {
		ctx,cancel := context.WithTimeout(context.Background(),client.Timeout)
		
		req,err := http.NewRequestWithContext(ctx,"HEAD",url,nil)
		if err != nil {
			results <- Result{Url: url,Error: err.Error()}
			log.Printf("error in process request %v",err)
			cancel()
			continue
		}
		resp,err := client.Do(req)
		if err != nil {
			results <- Result{Url: url,Error: err.Error()}
			log.Printf("error in process request %v",err)
			cancel()
			continue
		}
		results <- Result{Url: url,Status: resp.StatusCode}
		resp.Body.Close()
		cancel()

	}
}

func Writer(results <-chan Result, outputfile *os.File){
	writer := bufio.NewWriter(outputfile)
	for result := range results {
		if result.Error != ""{
			fmt.Fprintf(writer,"%s ERROR : %s\n",result.Url,result.Error)
		}else {
			fmt.Fprintf(writer,"%s %d\n",result.Url,result.Status)
		}
	}
	//important : change the remaining buffer to disk
	writer.Flush()
}
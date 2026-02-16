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
)

// Now i am explaing how i want to do this task but you read my approach and then talk to me is it worakble in production and industry level and a good architect so first
// first i make the file in h=which there is 100k urls then i created a queue(channel) of 1000 size and make the 500 workers. queue will take th eurls from file and and workers(goroutime )
// process these urls and check the status of urls using the head[i think this is a working pool] and every worker after knowing the status, update the file with status .can this apporach is good and work in production and becakedn archetect level?
// explain the existance of this approach
type Result struct {
	Url string
	Status int
	Error string

}

const (
	workerCount = 10
	queueSize = 100
	Timeout = 5 * time.Second
)


func main(){
	inputfile,err := os.Open("urls.txt")
	if err!=nil {
		panic(err)
	}
	defer inputfile.Close()

	//create the output file 
	outputfile,err := os.Create("results.txt")
	if err != nil {
		panic(err)
	}
	defer outputfile.Close()
	//created the queues
	jobs := make(chan string,queueSize)
	results := make(chan Result,queueSize)

	client := &http.Client{
		Timeout : Timeout,
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
		ctx,cancel := context.WithTimeout(context.Background(),Timeout)
		
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
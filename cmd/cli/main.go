package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type StressTestParams struct {
	Url          string
	RequestCount int64
	Concurrency  int64
}

type ReportRequest struct {
	StatusOk     int
	StatusOther  int
	ErrorRequest int
}

func main() {
	params := getParams()
	fmt.Println("Starting requests to:", params.Url)
	reportChannel := make(chan ReportRequest, 1)

	totalOK := 0
	totalOther := 0
	totalError := 0

	var wgChannel sync.WaitGroup
	wgChannel.Add(1)
	go func() {
		defer wgChannel.Done()
		for report := range reportChannel {
			totalOK += report.StatusOk
			totalOther += report.StatusOther
			totalError += report.ErrorRequest

		}
	}()

	totalRequest := params.RequestCount / params.Concurrency
	var wg sync.WaitGroup
	start := time.Now()
	for i := int64(0); i < params.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			makeRequest(params.Url, totalRequest, reportChannel)
		}()
	}
	wg.Wait()
	close(reportChannel)
	wgChannel.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Elapsed time: %s\n", elapsed)
	fmt.Println("Total requests:", (totalOK + totalOther + totalError))
	fmt.Println("Total HTTP 200 response: ", totalOK)
	fmt.Println("Other responses: ", totalOther)
	fmt.Println("Error requests: ", totalError)

}

func makeRequest(u string, requests int64, reportChannel chan<- ReportRequest) {
	_, err := url.ParseRequestURI(u)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return
	}

	requestOKCount := 0
	requestOtherCount := 0
	requestErrorCount := 0

	for i := 0; i < int(requests); i++ {
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			requestErrorCount++
			continue
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			// fmt.Println("Error making request:", err)
			requestErrorCount++
		} else {
			if resp.StatusCode == http.StatusOK {
				requestOKCount++
			} else {
				requestOtherCount++
			}
			defer resp.Body.Close()
		}
	}
	reportChannel <- ReportRequest{
		StatusOk:     requestOKCount,
		StatusOther:  requestOtherCount,
		ErrorRequest: requestErrorCount,
	}
}

func getParams() *StressTestParams {
	var params StressTestParams
	params = *getUserFlags()
	if params.Url == "" {
		params = *getUserInput()
	}

	if params.Url == "" {
		fmt.Println("URL is required")
		os.Exit(1)
	}
	if params.RequestCount <= 0 {
		params.RequestCount = 100
	}
	if params.Concurrency == 0 {
		params.Concurrency = 5
	}
	return &params
}

func getUserFlags() *StressTestParams {
	url := flag.String("url", "", "URL to stress test")
	requests := flag.String("requests", "", "Number of requests to send")
	concurrency := flag.String("concurrency", "", "Number of concurrent requests")

	flag.Parse()

	requestCount, _ := strconv.ParseInt(*requests, 10, 0)
	concurrencyCount, _ := strconv.ParseInt(*concurrency, 10, 64)

	return &StressTestParams{
		Url:          *url,
		RequestCount: requestCount,
		Concurrency:  concurrencyCount,
	}
}

func getUserInput() *StressTestParams {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter URL: ")
	url, _ := reader.ReadString('\n')

	fmt.Print("Enter Request Count: ")
	requestCountStr, _ := reader.ReadString('\n')
	requestCountStr = strings.TrimSpace(requestCountStr)
	requestCount, _ := strconv.ParseInt(requestCountStr, 10, 0)

	fmt.Print("Enter Concurrency: ")
	concurrencyStr, _ := reader.ReadString('\n')
	concurrencyStr = strings.TrimSpace(concurrencyStr)
	concurrency, _ := strconv.ParseInt(concurrencyStr, 10, 64)

	return &StressTestParams{
		Url:          url,
		RequestCount: requestCount,
		Concurrency:  concurrency,
	}
}

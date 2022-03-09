package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	parallelFlag = flag.Int("parallel", 10, "parallel processing value")
)

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] url1 url2 ...\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	urls := flag.Args()
	if len(urls) == 0 {
		fmt.Println("no urls provided")
		os.Exit(1)
	}

	var (
		wg       sync.WaitGroup
		inChan   = make(chan string)
		outChan  = make(chan [2]string)
		exitChan = make(chan struct{})
	)

	for i := 0; i < *parallelFlag; i++ {
		wg.Add(1)
		go processUrl(&wg, inChan, outChan)
	}
	go func() {
		for _, v := range urls {
			inChan <- prepareUrl(v)
		}
		close(inChan)
	}()

	go func() {
		for v := range outChan {
			fmt.Println(v[0], v[1])
		}
		exitChan <- struct{}{}
	}()

	wg.Wait()
	close(outChan)
	<-exitChan
}

// prepareUrl checks presented url schema
// if schema is not presented, http:// prefix is being added
func prepareUrl(url string) string {
	if !strings.HasPrefix(url, "http") {
		return "http://" + url
	}
	return url
}

// processUrl processes channels of urls concurrently
func processUrl(wg *sync.WaitGroup, in <-chan string, out chan<- [2]string) {
	defer wg.Done()
	var (
		hc  string
		err error
	)

	cli := &http.Client{}

	for url := range in {
		if hc, err = getHashedUrlContent(cli, url); err != nil {
			continue
		}
		out <- [2]string{url, hc}
	}
}

// getHashedUrlContent returns hashed response of presented url
func getHashedUrlContent(cli *http.Client, url string) (string, error) {
	h := md5.New()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("get request: %v", err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		return "", fmt.Errorf("http get: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if _, err = io.Copy(h, resp.Body); err != nil {
		return "", fmt.Errorf("io copy: %v", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

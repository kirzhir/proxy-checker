package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/kirzhir/proxy-checker/pkg/checker"
)

type Configuration struct {
	CheckerEndpount   string
	ConnectionTimeout uint
}

type Result struct {
	alive bool
	proxy string
}

func main() {

	proxyListFilepath := flag.String("proxy-list-path", "", "required param proxy-lsit filepath")
	configFile := flag.String("config", "", "required param config filepath")

	flag.Parse()

	configFilepath, err := filepath.Abs(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	config, err := initConfiguration(configFilepath)
	if err != nil {
		log.Fatal(err)
	}

	if *proxyListFilepath == "" {
		log.Fatalf("Missing required param proxy-list-path")
	}

	if _, err := os.Stat(*proxyListFilepath); os.IsNotExist(err) {
		log.Fatalf("File: \"%v\" doesn`t exist", *proxyListFilepath)
	}

	if ext := filepath.Ext(*proxyListFilepath); ext != ".txt" {
		log.Fatalf("Unsupported filetype: \"%v\"", ext)
	}

	file, err := os.Open(*proxyListFilepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var ok bool
	var proxy string
	regexpcmp, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]):[0-9]+$`)

	jobs := make(chan string)
	results := make(chan *Result)

	wg := &sync.WaitGroup{}
	c := checker.NewChecker(config.CheckerEndpount, config.ConnectionTimeout)
	for w := 1; w <= 10; w++ {
		wg.Add(1)
		go func(id int, c *checker.Checker, jobs <-chan string, results chan<- *Result) {
			defer wg.Done()
			for proxy := range jobs {
				alive, err := c.Check(proxy)
				if err != nil {
					// log.Println(err)
				}

				results <- &Result{alive, proxy}
			}
		}(w, c, jobs, results)
	}

	scanner := bufio.NewScanner(file)
	go func() {
		for scanner.Scan() {
			proxy = scanner.Text()
			if ok = regexpcmp.MatchString(proxy); !ok {
				continue
			}

			jobs <- proxy
		}
		close(jobs)

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}()

	var alive []string
	exit := make(chan interface{})
	go func() {
		for result := range results {
			if !result.alive {
				fmt.Printf("%s failed\n", result.proxy)
			} else {
				alive = append(alive, result.proxy)
			}
		}

		exit <- struct{}{}
	}()

	wg.Wait()
	close(results)

	<-exit
	for proxy := range alive {
		fmt.Println(proxy)
	}
}

func initConfiguration(configFilepath string) (*Configuration, error) {

	if configFilepath == "" {
		return nil, fmt.Errorf("Empty config filename: %v", configFilepath)
	}

	file, err := os.Open(configFilepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	configuration := Configuration{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		return nil, err
	}

	return &configuration, nil
}

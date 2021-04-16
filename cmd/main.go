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

	"github.com/kirzhir/proxy-checker/pkg/checker"
)

type Configuration struct {
	CheckerEndpount   string
	ConnectionTimeout uint
}

func main() {

	proxyListFilepath := flag.String("proxy-list-path", "", "required param proxy-lsit filepath")
	configFilepath := flag.String("config-path", "", "required param config filepath")
	flag.Parse()

	config, err := initConfiguration(*configFilepath)
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

	c := checker.NewChecker(config.CheckerEndpount, config.ConnectionTimeout)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxy = scanner.Text()
		if ok = regexpcmp.MatchString(proxy); !ok {
			continue
		}

		_, err := c.Check(proxy)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Alive: %v", proxy)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
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
